package servicemessages

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	exp_processor_mock "gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_processor/mocks"
	msgmocks "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages/mocks"
	report_requester_mock "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester/mocks"
)

var currencies = map[string]struct{}{
	"USD": {},
	"RUB": {},
	"EUR": {},
	"CNY": {},
}

func TestOnStartCommandShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := msgmocks.NewMockMessageSender(ctrl)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	model := New(sender, currencies, processor, reportRequester, nil, nil)
	ctx := context.Background()
	userId := int64(100)

	msg := "Привет, я буду считать твои деньги. Вот что я умею:\n" +
		"addExpense- добавить трату\n" +
		"Пример: /addExpense 10;Дом;2022-10-04 10:00:00\n" +
		"getExpenses - получить список трат за неделю, месяц и год\n" +
		"Пример: /getExpenses week\n" +
		"requestCurrencyChange - вызвать менюсмены валюты\n" +
		"setCurrency - установить валюту ввода и отображения отчетов.\n" +
		"Пример: /setCurrency EUR\n" +
		"setLimit - установить лимит трат на категорию.\n" +
		"Пример: /setLimit Ремонт 1200.50\n"

	sender.EXPECT().SendMessage(msg, userId, mainMenu)

	err := model.IncomingMessage(ctx, Message{
		Command: startCommand,
		UserID:  userId,
	})

	assert.NoError(t, err)
}

func TestOnUnknownCommandShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123), mainMenu)
	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor.EXPECT().AddExpense(wrapedCtx, 125.5, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(wrapedCtx, "Кофе", "RUB", userId)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitSetShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Свободный месячный лимит 10.00 RUB",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	processor.EXPECT().AddExpense(wrapedCtx, 125.50, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(wrapedCtx, "Кофе", "RUB", userId).Return(10.00, true, nil)

	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitReachedShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Достигнут месячный лимит (-12.00 RUB)",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	processor.EXPECT().AddExpense(wrapedCtx, 125.50, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(wrapedCtx, "Кофе", "RUB", userId).Return(-12.00, true, nil)

	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверное количество параметров.\n"+
		"Ожидается: Сумма;Категория;Дата \n"+
		"Например: 120.50;Дом;2022-10-01 13:25:23", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command: addExpenseCommand,
		UserID:  userId,
	})

	assert.NoError(t, err)
}

func TestOnGetWeekExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Запрос на формирование отчета отправлен", userId, mainMenu)

	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	reportRequester.EXPECT().SendRequestReport(wrapedCtx, userId, model.Week, "RUB")

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command: getExpensesCommand,
		UserID:  userId,
	})

	assert.NoError(t, err)
}

func TestOnGetMonthExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Запрос на формирование отчета отправлен", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)

	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	reportRequester.EXPECT().SendRequestReport(wrapedCtx, userId, model.Month, "RUB")

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "month",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnGetYearExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Запрос на формирование отчета отправлен", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)

	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)
	reportRequester.EXPECT().SendRequestReport(wrapedCtx, userId, model.Year, "RUB")

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "year",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnGetExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверный период. Ожидается: year, month, week. По-умолчанию week", int64(123), mainMenu)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "wrong",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnRequestCurrencyChangeShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage(
		"Выберите валюту",
		int64(123),
		[]string{"/setCurrency CNY", "/setCurrency EUR", "/setCurrency RUB", "/setCurrency USD"},
	)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          requestCurrencyChangeCommand,
		CommandArguments: "",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnSetCurrenctShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage(
		"Установлена валюта в USD",
		int64(123),
		mainMenu,
	)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setCurrencyCommand,
		CommandArguments: "USD",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnSetCurrenctShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неизвестная валюта FOO", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setCurrencyCommand,
		CommandArguments: "FOO",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnSetLimitShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверное количество параметров.\nОжидается: Категория;Сумма \nНапример: Дом;12000.50", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "invalid",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnSetLimitShouldAnswerWithSuccessMessage(t *testing.T) {
	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")
	_, wrapedCtx = opentracing.StartSpanFromContext(wrapedCtx, "wrap2")

	userId := int64(100)
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Установлен месячный лимит 12500.50 RUB для категории Дом", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reportRequester := report_requester_mock.NewMockReportRequester(ctrl)

	processor.EXPECT().SetLimit(wrapedCtx, "Дом", userId, 12500.50, "RUB").Return(12500.50, nil)

	model := New(sender, currencies, processor, reportRequester, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "Дом;12500.50",
		UserID:           userId,
	})

	assert.NoError(t, err)
}
