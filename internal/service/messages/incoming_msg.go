package servicemessages

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/expenses"
)

const (
	errAddExpenseInvalidParameterMessage = "неверное количество параметров.\nОжидается: Сумма;Категория;Дата \n" +
		"Например: 120.50;Дом;2022-10-01 13:25:23"
	errInvalidAmountParameterMessage             = "неверное значение суммы: %v"
	errAddExpenseInvalidDatetimeParameterMessage = "неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11"
	errSaveExpenseMessage                        = "ошибка сохранения траты"
	errUnknownCurrency                           = "неизвестная валюта %s"
	errGetExpensesInvalidPeriodMessage           = "неверный период. Ожидается: year, month, week. По-умолчанию week"
	errSetLimitInvalidParameterMessage           = "неверное количество параметров.\nОжидается: Категория;Сумма \n" +
		"Например: Дом;12000.50"
	msgExpenseAdded = "Трата %.02f %s добавлена в категорию %s с датой %s"
	msgCurrencySet  = "Установлена валюта в %s"
	msgFreeLimit    = "Свободный месячный лимит %.02f %s"
	msgLimitReached = "Достигнут месячный лимит (%.02f %s)"
	msgSetLimit     = "Установлен месячный лимит %.02f %s для категории %s"

	datetimeFormat = "2006-01-02 15:04:05"

	startCommand                 = "start"
	addExpenseCommand            = "addExpense"
	getExpensesCommand           = "getExpenses"
	requestCurrencyChangeCommand = "requestCurrencyChange"
	setCurrencyCommand           = "setCurrency"
	setLimitCommand              = "setLimit"

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
	storage   repo.ExpensesRepository
	converter serviceconverter.Converter
	currency  string
}

func New(tgClient MessageSender, storage repo.ExpensesRepository, conv serviceconverter.Converter) *Model {
	return &Model{
		tgClient:  tgClient,
		storage:   storage,
		converter: conv,
		currency:  serviceconverter.RUB,
	}
}

type Message struct {
	Command          string
	CommandArguments string
	Text             string
	UserID           int64
}

func (m *Model) IncomingMessage(ctx context.Context, msg Message) error {
	response := "не знаю эту команду"
	var err error
	btns := mainMenu

	switch msg.Command {
	case startCommand:
		response = m.showInfo()
	case addExpenseCommand:
		response, err = m.addExpense(ctx, msg)
	case getExpensesCommand:
		response, err = m.getExpenses(ctx, msg)
	case requestCurrencyChangeCommand:
		response, btns = m.requestCurrencyChange()
	case setCurrencyCommand:
		response, err = m.setCurrency(msg)
	case setLimitCommand:
		response, err = m.setLimit(ctx, msg)
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
		setLimitCommand,
		" - установить лимит трат на категорию.\nПример: /setLimit Ремонт 1200.50\n",
	}, "")
}

func (m *Model) addExpense(ctx context.Context, msg Message) (string, error) {
	parts := strings.Split(msg.CommandArguments, ";")

	if len(parts) != 3 {
		return "", errors.New(errAddExpenseInvalidParameterMessage)
	}

	trimmedAmount := strings.Trim(parts[0], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return "", fmt.Errorf(errInvalidAmountParameterMessage, trimmedAmount)
	}

	trimmedDatetime := strings.Trim(parts[2], " ")
	datetime, err := time.Parse(datetimeFormat, trimmedDatetime)
	if err != nil {
		return "", fmt.Errorf(errAddExpenseInvalidDatetimeParameterMessage, trimmedDatetime)
	}

	convertedAmount := m.converter.ToRUB(amount, m.currency)

	trimmedCategory := strings.Trim(parts[1], " ")
	err = m.storage.Add(ctx, expenses.Expense{
		Amount:   int64(convertedAmount * primitiveCurrencyMultiplier),
		Category: trimmedCategory,
		Datetime: datetime,
	})
	if err != nil {
		return "", errors.Wrap(err, errSaveExpenseMessage)
	}

	freeLimit, hasLimit, err := m.storage.GetFreeLimit(ctx, trimmedCategory)
	if err != nil {
		return "", errors.Wrap(err, errSaveExpenseMessage)
	}
	var responseMsg string
	responseMsg = msgExpenseAdded

	if hasLimit {
		var addMsg string
		convertedFreeLimit := m.converter.FromRUB(float64(freeLimit), m.currency)
		if freeLimit > 0 {
			addMsg = msgFreeLimit
		} else {
			addMsg = msgLimitReached
		}
		responseMsg = fmt.Sprintf(
			"%s.\n%s", responseMsg,
			fmt.Sprintf(addMsg, convertedFreeLimit/100, m.currency),
		)
	}

	return fmt.Sprintf(responseMsg, amount, m.currency, trimmedCategory, trimmedDatetime), nil
}

func (m *Model) getExpenses(ctx context.Context, msg Message) (string, error) {
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

	expenses, err := m.storage.GetExpenses(ctx, expPeriod)
	if err != nil {
		return "", err
	}

	result := make(map[string]int64) // [категория]сумма

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

	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i] < currencies[j]
	})

	return "Выберите валюту", currencies
}

func (m *Model) setCurrency(msg Message) (string, error) {
	currencies := m.converter.GetAvailableCurrencies()

	if _, found := currencies[msg.CommandArguments]; !found {
		return "", fmt.Errorf(errUnknownCurrency, msg.CommandArguments)
	}

	m.currency = msg.CommandArguments

	return fmt.Sprintf(msgCurrencySet, msg.CommandArguments), nil
}

func (m *Model) setLimit(ctx context.Context, msg Message) (string, error) {
	parts := strings.Split(msg.CommandArguments, ";")

	if len(parts) != 2 {
		return "", errors.New(errSetLimitInvalidParameterMessage)
	}

	trimmedCategory := strings.Trim(parts[0], " ")

	trimmedAmount := strings.Trim(parts[1], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return "", fmt.Errorf(errInvalidAmountParameterMessage, trimmedAmount)
	}

	convertedAmount := m.converter.ToRUB(amount, m.currency)
	if err := m.storage.SetLimit(ctx, trimmedCategory, int64(convertedAmount*primitiveCurrencyMultiplier)); err != nil {
		return "", err
	}
	return fmt.Sprintf(msgSetLimit, convertedAmount, m.currency, trimmedCategory), nil
}
