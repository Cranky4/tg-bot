package servicemessages

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	expense_service "gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense"
)

const (
	errAddExpenseInvalidParameterMessage = "неверное количество параметров.\nОжидается: Сумма;Категория;Дата \n" +
		"Например: 120.50;Дом;2022-10-01 13:25:23"
	errInvalidAmountParameterMessage             = "неверное значение суммы: %v"
	errAddExpenseInvalidDatetimeParameterMessage = "неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11"

	errUnknownCurrency                 = "неизвестная валюта %s"
	errGetExpensesInvalidPeriodMessage = "неверный период. Ожидается: year, month, week. По-умолчанию week"
	errSetLimitInvalidParameterMessage = "неверное количество параметров.\nОжидается: Категория;Сумма \n" +
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
)

var mainMenu = []string{
	strings.Join([]string{"/", getExpensesCommand}, ""),
	strings.Join([]string{"/", requestCurrencyChangeCommand}, ""),
}

type MessageSender interface {
	SendMessage(text string, userID int64, buttons []string) error
}

type Model struct {
	tgClient         MessageSender
	currencies       map[string]struct{}
	expenseProcessor expense_service.ExpenseProcessor
	expenseReporter  expense_service.ExpenseReporter
	currency         string
}

func New(
	tgClient MessageSender,
	currencies map[string]struct{},
	expenseProcessor expense_service.ExpenseProcessor,
	expenseReporter expense_service.ExpenseReporter,
) *Model {
	return &Model{
		tgClient:         tgClient,
		currencies:       currencies,
		currency:         serviceconverter.RUB,
		expenseProcessor: expenseProcessor,
		expenseReporter:  expenseReporter,
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

	trimmedCategory := strings.Trim(parts[1], " ")

	if _, err = m.expenseProcessor.AddExpense(ctx, amount, m.currency, trimmedCategory, datetime); err != nil {
		return "", err
	}

	freeLimit, hasLimit, err := m.expenseProcessor.GetFreeLimit(ctx, trimmedCategory, m.currency)
	if err != nil {
		return "", err
	}

	var responseMsg string
	responseMsg = msgExpenseAdded

	if hasLimit {
		var addMsg string

		if freeLimit > 0 {
			addMsg = msgFreeLimit
		} else {
			addMsg = msgLimitReached
		}
		responseMsg = fmt.Sprintf(
			"%s.\n%s", responseMsg,
			fmt.Sprintf(addMsg, freeLimit, m.currency),
		)
	}

	return fmt.Sprintf(responseMsg, amount, m.currency, trimmedCategory, trimmedDatetime), nil
}

func (m *Model) getExpenses(ctx context.Context, msg Message) (string, error) {
	var expPeriod model.ExpensePeriod

	switch msg.CommandArguments {
	case "week":
		expPeriod = model.Week
	case "month":
		expPeriod = model.Month
	case "year":
		expPeriod = model.Year
	default:
		if msg.CommandArguments != "" {
			return "", errors.New(errGetExpensesInvalidPeriodMessage)
		}
		expPeriod = model.Week
	}

	report, err := m.expenseReporter.GetReport(ctx, expPeriod, m.currency)
	if err != nil {
		return "", err
	}

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", &expPeriod),
	)
	defer reporter.Reset()

	if report.IsEmpty {
		reporter.WriteString("пусто\n")
	}

	for category, amount := range report.Rows {
		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02f %s\n", category, amount, m.currency)); err != nil {
			return "", err
		}
	}

	return reporter.String(), nil
}

func (m *Model) requestCurrencyChange() (string, []string) {
	currencies := make([]string, 0, len(m.currencies))
	for c := range m.currencies {
		currencies = append(currencies, strings.Join([]string{"/", setCurrencyCommand, " ", c}, ""))
	}

	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i] < currencies[j]
	})

	return "Выберите валюту", currencies
}

func (m *Model) setCurrency(msg Message) (string, error) {
	if _, found := m.currencies[msg.CommandArguments]; !found {
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

	convertedAmount, err := m.expenseProcessor.SetLimit(ctx, trimmedCategory, amount, m.currency)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(msgSetLimit, convertedAmount, m.currency, trimmedCategory), nil
}
