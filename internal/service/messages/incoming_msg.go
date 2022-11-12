package servicemessages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber/jaeger-client-go"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_processor"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequester "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester"
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
	tgClient             MessageSender
	currencies           map[string]struct{}
	expenseProcessor     expense_processor.ExpenseProcessor
	reportRequester      reportrequester.ReportRequester
	currency             string
	totalRequestsCounter *prometheus.CounterVec
	responseTimeSummary  *prometheus.SummaryVec
}

func New(
	tgClient MessageSender,
	currencies map[string]struct{},
	expenseProcessor expense_processor.ExpenseProcessor,
	reportRequester reportrequester.ReportRequester,
	totalRequestsCounter *prometheus.CounterVec,
	responseTimeSummary *prometheus.SummaryVec,
) *Model {
	return &Model{
		tgClient:             tgClient,
		currencies:           currencies,
		currency:             serviceconverter.RUB,
		expenseProcessor:     expenseProcessor,
		reportRequester:      reportRequester,
		totalRequestsCounter: totalRequestsCounter,
		responseTimeSummary:  responseTimeSummary,
	}
}

type Message struct {
	Command          string
	CommandArguments string
	Text             string
	UserID           int64
}

func (m *Model) IncomingMessage(ctx context.Context, msg Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Messaging_IncomingMessage")
	defer span.Finish()

	// Меняет ид трейса для логов
	if spanCtx, ok := span.Context().(jaeger.SpanContext); ok {
		logger.SetTraceId(spanCtx.TraceID().String())
	}

	// Метрика времени ответа
	if m.responseTimeSummary != nil {
		start := time.Now()
		defer func(start time.Time, command string) {
			m.responseTimeSummary.WithLabelValues(command).Observe(float64(time.Since(start).Milliseconds()))
		}(start, msg.Command)
	}

	// Метрика количества команд
	if m.totalRequestsCounter != nil {
		m.totalRequestsCounter.WithLabelValues(msg.Command).Inc()
	}

	logger.Debug(
		"получена команда",
		logger.LogDataItem{Key: "userId", Value: msg.UserID},
		logger.LogDataItem{Key: "command", Value: msg.Command},
		logger.LogDataItem{Key: "arguments", Value: msg.CommandArguments},
	)

	response := "не знаю эту команду"
	var err error
	btns := mainMenu

	switch msg.Command {
	case startCommand:
		response = m.showInfo(ctx)
	case addExpenseCommand:
		response, err = m.addExpense(ctx, msg)
	case getExpensesCommand:
		response, err = m.getExpenses(ctx, msg)
	case requestCurrencyChangeCommand:
		response, btns = m.requestCurrencyChange(ctx)
	case setCurrencyCommand:
		response, err = m.setCurrency(ctx, msg)
	case setLimitCommand:
		response, err = m.setLimit(ctx, msg)
	}

	if err != nil {
		response = err.Error()

		logger.Error(response)
	}

	return m.tgClient.SendMessage(response, msg.UserID, btns)
}

func (m *Model) SendReport(ctx context.Context, report *expense_reporter.ExpenseReport) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Messaging_SendReport")
	defer span.Finish()

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", report.Period.String()),
	)
	defer reporter.Reset()

	if report.IsEmpty() {
		reporter.WriteString("пусто\n")
	}

	for category, amount := range report.Rows {
		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02f %s\n", category, amount, m.currency)); err != nil {
			return err
		}
	}

	return m.tgClient.SendMessage(reporter.String(), report.UserID, mainMenu)
}
