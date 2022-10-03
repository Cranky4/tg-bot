package messages

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/storage"
)

const (
	errAddExpenseInvalidParameterMessage         = "Неверное количество параметров.\nОжидается: Сумма;Категория;Дата \nНапример: 120.50;Дом;2022-10-01 13:25:23"
	errAddExpenseInvalidAmountParameterMessage   = "Неверное значение суммы: %v"
	errAddExpenseInvalidDatetimeParameterMessage = "Неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11"
	errSaveExpenseMessage                        = "Ошибка сохранения траты"

	errGetExpensesInvalidPeriodMessage = "Неверный период. Ожидается: year, month, week. По-умолчанию week"

	msgExpenseAdded = "Трата %.02fр добавлена в категорию %s с датой %s"

	datetimeFormat = "2006-01-02 15:04:05"

	startCommand       = "start"
	addExpenseCommand  = "addExpense"
	getExpensesCommand = "getExpenses"
)

type MessageSender interface {
	SendMessage(text string, userID int64) error
}

type Model struct {
	tgClient MessageSender
	storage  storage.Storage
}

func New(tgClient MessageSender, storage storage.Storage) *Model {
	return &Model{
		tgClient: tgClient,
		storage:  storage,
	}
}

type Message struct {
	Command          string
	CommandArguments string
	Text             string
	UserID           int64
}

func (m *Model) IncomingMessage(msg Message) error {
	response := "не знаю эту команду"
	var err error

	switch msg.Command {
	case startCommand:
		response = "hello"
	case addExpenseCommand:
		response, err = m.addExpense(msg)
	case getExpensesCommand:
		response, err = m.getExpenses(msg)
	}

	if err != nil {
		response = err.Error()
	}

	return m.tgClient.SendMessage(response, msg.UserID)
}

func (m *Model) addExpense(msg Message) (string, error) {
	parts := strings.Split(msg.CommandArguments, ";")
	var responseMsg string

	if len(parts) != 3 {
		return responseMsg, errors.New(errAddExpenseInvalidParameterMessage)
	}

	trimmedAmount := strings.Trim(parts[0], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return responseMsg, fmt.Errorf(errAddExpenseInvalidAmountParameterMessage, trimmedAmount)
	}

	trimmedDatetime := strings.Trim(parts[2], " ")
	datetime, err := time.Parse(datetimeFormat, trimmedDatetime)
	if err != nil {
		return responseMsg, fmt.Errorf(errAddExpenseInvalidDatetimeParameterMessage, trimmedDatetime)
	}

	trimmedCategory := strings.Trim(parts[1], " ")
	err = m.storage.Add(expenses.Expense{
		Amount:   int(float32(amount) * 100),
		Category: trimmedCategory,
		Datetime: datetime,
	})
	if err != nil {
		return responseMsg, errors.New(errSaveExpenseMessage)
	}

	return fmt.Sprintf(msgExpenseAdded, amount, trimmedCategory, trimmedDatetime), nil
}

func (m *Model) getExpenses(msg Message) (string, error) {
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
			return "", errors.New(errGetExpensesInvalidPeriodMessage)
		}
		expPeriod = expenses.Week
	}

	expenses := m.storage.GetExpenses(expPeriod)

	result := make(map[string]int) // [категория]сумма

	for _, e := range expenses {
		result[e.Category] += e.Amount
	}

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", &expPeriod),
	)
	defer reporter.Reset()

	if len(result) == 0 {
		reporter.WriteString("пусто\n")
	}

	for category, amount := range result {
		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02fр\n", category, float32(amount)/100)); err != nil {
			return "", err
		}
	}

	return reporter.String(), nil
}
