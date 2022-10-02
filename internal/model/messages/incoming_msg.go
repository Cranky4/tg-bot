package messages

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

type MsgError struct {
	message string
}

func (e MsgError) Error() string {
	return e.message
}

const (
	errAddExpenseInvalidParameterMessage         string = "Неверное количество параметров.\nОжидается: Сумма;Категория;Дата \nНапример: 120.50;Дом;2022-10-01 13:25:23"
	errAddExpenseInvalidAmountParameterMessage   string = "Неверное значение суммы: %v"
	errAddExpenseInvalidDatetimeParameterMessage string = "Неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11"
	errSaveExpenseMessage                        string = "Ошибка сохранения траты"

	errGetExpensesInvalidPeriodMessage string = "Неверный период. Ожидается: year, month, week. По-умолчанию week"

	msgExpenseAdded string = "Трата %.02fр добавлена в категорию %s с датой %s"

	datetimeFormat string = "2006-01-02 15:04:05"
)

type MessageSender interface {
	SendMessage(text string, userID int64) error
}

type Storage interface {
	Add(expense expenses.Expense) error
	GetExpenses(expense expenses.ExpensePeriod) []*expenses.Expense
}

type Model struct {
	tgClient MessageSender
	storage  Storage
}

func New(tgClient MessageSender, storage Storage) *Model {
	return &Model{
		tgClient: tgClient,
		storage:  storage,
	}
}

type Message struct {
	Command, CommandArguments, Text string
	UserID                          int64
}

const (
	StartCommand       string = "start"
	AddExpenseCommand  string = "addExpense"
	GetExpensesCommand string = "getExpenses"
)

func (m *Model) IncomingMessage(msg Message) error {
	response := "не знаю эту команду"
	var err error

	switch msg.Command {
	case StartCommand:
		response = "hello"
	case AddExpenseCommand:
		response, err = m.addExpense(msg)
	case GetExpensesCommand:
		response, err = m.getExpenses(msg)
	}

	if err != nil {
		return m.tgClient.SendMessage(err.Error(), msg.UserID)
	}

	return m.tgClient.SendMessage(response, msg.UserID)
}

func (s *Model) addExpense(msg Message) (responseMsg string, err error) {
	parts := strings.Split(msg.CommandArguments, ";")

	if len(parts) != 3 {
		err = MsgError{message: errAddExpenseInvalidParameterMessage}
		return
	}

	trimmedAmount := strings.Trim(parts[0], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		err = MsgError{message: fmt.Sprintf(errAddExpenseInvalidAmountParameterMessage, trimmedAmount)}
		return
	}

	trimmedDatetime := strings.Trim(parts[2], " ")
	datetime, err := time.Parse(datetimeFormat, trimmedDatetime)
	if err != nil {
		err = MsgError{message: fmt.Sprintf(errAddExpenseInvalidDatetimeParameterMessage, trimmedDatetime)}
		return
	}

	trimmedCategory := strings.Trim(parts[1], " ")
	err = s.storage.Add(expenses.Expense{
		Amount:   int(float32(amount) * 100),
		Category: trimmedCategory,
		Datetime: datetime,
	})
	if err != nil {
		err = MsgError{message: errSaveExpenseMessage}
		return
	}

	responseMsg = fmt.Sprintf(msgExpenseAdded, amount, trimmedCategory, trimmedDatetime)
	return
}

func (s *Model) getExpenses(msg Message) (responseMsg string, err error) {
	var expPeriod expenses.ExpensePeriod

	switch msg.CommandArguments {
	case "week":
		expPeriod = expenses.Week
	case "month":
		expPeriod = expenses.Month
	case "year":
		expPeriod = expenses.Year
	default:
		if msg.CommandArguments != "" {
			err = MsgError{message: errGetExpensesInvalidPeriodMessage}
			return
		}
		expPeriod = expenses.Week
	}

	expenses := s.storage.GetExpenses(expPeriod)

	result := make(map[string]int) // [категория]сумма

	for _, e := range expenses {
		result[e.Category] += e.Amount
	}

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", expPeriod),
	)
	defer reporter.Reset()

	if len(result) == 0 {
		reporter.WriteString("пусто\n")
	}

	for category, amount := range result {
		reporter.WriteString(
			fmt.Sprintf("%s - %.02fр\n", category, float32(amount)/100),
		)
	}

	responseMsg = reporter.String()
	return
}
