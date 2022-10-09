package messages

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

const (
	errAddExpenseInvalidParameterMessage = "неверное количество параметров.\nОжидается: Сумма;Категория;Дата \n" +
		"Например: 120.50;Дом;2022-10-01 13:25:23"
	errAddExpenseInvalidAmountParameterMessage   = "неверное значение суммы: %v"
	errAddExpenseInvalidDatetimeParameterMessage = "неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11"
	errSaveExpenseMessage                        = "ошибка сохранения траты"

	errGetExpensesInvalidPeriodMessage = "неверный период. Ожидается: year, month, week. По-умолчанию week"

	msgExpenseAdded = "Трата %.02f %s добавлена в категорию %s с датой %s"

	datetimeFormat = "2006-01-02 15:04:05"

	startCommand                 = "start"
	addExpenseCommand            = "addExpense"
	getExpensesCommand           = "getExpenses"
	requestCurrencyChangeCommand = "requestCurrencyChange"
	setCurrencyCommand           = "setCurrency"

	primitiveCurrencyMultiplier = 100
)

var mainMenu = []string{
	strings.Join([]string{"/", getExpensesCommand}, ""),
	strings.Join([]string{"/", requestCurrencyChangeCommand}, ""),
}

type MessageSender interface {
	SendMessage(text string, userID int64, buttons []string) error
}

type Model struct {
	tgClient  MessageSender
	storage   Storage
	converter converter.Converter
	currency  string
}

func New(tgClient MessageSender, storage Storage, conv converter.Converter) *Model {
	return &Model{
		tgClient:  tgClient,
		storage:   storage,
		converter: conv,
		currency:  converter.RUB,
	}
}

type Storage interface {
	Add(expense expenses.Expense) error
	GetExpenses(period expenses.ExpensePeriod) ([]*expenses.Expense, error)
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
	btns := mainMenu

	switch msg.Command {
	case startCommand:
		response = m.showInfo()
	case addExpenseCommand:
		response, err = m.addExpense(msg)
	case getExpensesCommand:
		response, err = m.getExpenses(msg)
	case requestCurrencyChangeCommand:
		response, btns = m.requestCurrencyChange()
	case setCurrencyCommand:
		response, err = m.setCurrency(msg)
	}

	if err != nil {
		response = err.Error()
	}

	return m.tgClient.SendMessage(response, msg.UserID, btns)
}

func (m *Model) showInfo() string {
	return strings.Join([]string{
		"Привет, я буду считать твои деньги. Вот что я умею:\n",
		addExpenseCommand,
		"- добавить трату\nПример: /addExpense 10;Дом;2022-10-04 10:00:00\n",
		getExpensesCommand,
		" - получить список трат за неделю, месяц и год\nПример: /getExpenses week\n",
		requestCurrencyChangeCommand,
		" - вызвать менюсмены валюты\n",
		setCurrencyCommand,
		" - установить валюту ввода и отображения отчетов.\nПример: /setCurrency EUR\n",
	}, "")
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

	convertedAmount := m.converter.ToRUB(amount, m.currency)

	trimmedCategory := strings.Trim(parts[1], " ")
	err = m.storage.Add(expenses.Expense{
		Amount:   int(convertedAmount * primitiveCurrencyMultiplier),
		Category: trimmedCategory,
		Datetime: datetime,
	})
	if err != nil {
		return responseMsg, errors.Wrap(err, errSaveExpenseMessage)
	}

	return fmt.Sprintf(msgExpenseAdded, amount, m.currency, trimmedCategory, trimmedDatetime), nil
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

	expenses, err := m.storage.GetExpenses(expPeriod)
	if err != nil {
		return "", err
	}

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
		converted := m.converter.FromRUB(float64(amount/primitiveCurrencyMultiplier), m.currency)

		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02f %s\n", category, converted, m.currency)); err != nil {
			return "", err
		}
	}

	return reporter.String(), nil
}

func (m *Model) requestCurrencyChange() (string, []string) {
	currs := m.converter.GetAvailableCurrencies()
	currencies := make([]string, 0, len(currs))
	for c := range m.converter.GetAvailableCurrencies() {
		currencies = append(currencies, strings.Join([]string{"/", setCurrencyCommand, " ", c}, ""))
	}

	return "Выберите валюту", currencies
}

func (m *Model) setCurrency(msg Message) (string, error) {
	currencies := m.converter.GetAvailableCurrencies()

	if _, found := currencies[msg.CommandArguments]; !found {
		return "", fmt.Errorf("неизвестная валюта %s", msg.CommandArguments)
	}

	m.currency = msg.CommandArguments

	return fmt.Sprintf("Установлена валюта в %s", msg.CommandArguments), nil
}
