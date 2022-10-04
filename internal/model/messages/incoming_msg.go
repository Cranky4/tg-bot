package messages

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/storage"
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
	strings.Join([]string{"/", addExpenseCommand}, ""),
	strings.Join([]string{"/", getExpensesCommand}, ""),
	strings.Join([]string{"/", requestCurrencyChangeCommand}, ""),
}

type MessageSender interface {
	SendMessage(text string, userID int64, buttons []string) error
}

type Model struct {
	tgClient  MessageSender
	storage   storage.Storage
	converter converter.Converter
	currency  converter.Currency
}

func New(tgClient MessageSender, storage storage.Storage, conv converter.Converter) *Model {
	return &Model{
		tgClient:  tgClient,
		storage:   storage,
		converter: conv,
		currency:  converter.RUB,
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
	btns := mainMenu

	switch msg.Command {
	case startCommand:
		response = "hello"
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
		return responseMsg, errors.New(errSaveExpenseMessage)
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
		converted := m.converter.FromRUB(float64(amount/primitiveCurrencyMultiplier), m.currency)

		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02f %s\n", category, converted, m.currency)); err != nil {
			return "", err
		}
	}

	return reporter.String(), nil
}

func (m *Model) requestCurrencyChange() (string, []string) {
	return "Выберите валюту", []string{
		strings.Join([]string{"/", setCurrencyCommand, " ", string(converter.USD)}, ""),
		strings.Join([]string{"/", setCurrencyCommand, " ", string(converter.EUR)}, ""),
		strings.Join([]string{"/", setCurrencyCommand, " ", string(converter.CNY)}, ""),
		strings.Join([]string{"/", setCurrencyCommand, " ", string(converter.RUB)}, ""),
	}
}

func (m *Model) setCurrency(msg Message) (string, error) {
	currencies := m.converter.GetAvailableCurrencies()

	_, curFound := currencies[converter.Currency(msg.CommandArguments)]

	if !curFound {
		return "", fmt.Errorf("неизвестная валюта %s", msg.CommandArguments)
	}

	m.currency = converter.Currency(msg.CommandArguments)

	return fmt.Sprintf("Установлена валюта в %s", msg.CommandArguments), nil
}
