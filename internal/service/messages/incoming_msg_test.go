package servicemessages

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	exp_processor_mock "gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_processor/mocks"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	exp_reporter_mock "gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter/mocks"
	service_logger "gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	msgmocks "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages/mocks"
)

var currencies = map[string]struct{}{
	"USD": {},
	"RUB": {},
	"EUR": {},
	"CNY": {},
}

type testLogger struct {
	logs []string
}

func (l *testLogger) Debug(msg string, data ...service_logger.LogDataItem) {
	l.logs = append(l.logs, fmt.Sprintf("[debug] %s (%v)", msg, data))
}

func (l *testLogger) Info(msg string, data ...service_logger.LogDataItem) {
	l.logs = append(l.logs, fmt.Sprintf("[info] %s (%v)", msg, data))
}

func (l *testLogger) Warn(msg string, data ...service_logger.LogDataItem) {
	l.logs = append(l.logs, fmt.Sprintf("[warn] %s (%v)", msg, data))
}

func (l *testLogger) Error(msg string, data ...service_logger.LogDataItem) {
	l.logs = append(l.logs, fmt.Sprintf("[error] %s (%v)", msg, data))
}

func (l *testLogger) Fatal(msg string, data ...service_logger.LogDataItem) {
	l.logs = append(l.logs, fmt.Sprintf("[error] %s (%v)", msg, data))
}

func TestOnStartCommandShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := msgmocks.NewMockMessageSender(ctrl)
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)
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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123), mainMenu)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor.EXPECT().AddExpense(ctx, 125.5, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(ctx, "Кофе", "RUB", userId)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitSetShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Свободный месячный лимит 10.00 RUB",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	processor.EXPECT().AddExpense(ctx, 125.50, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(ctx, "Кофе", "RUB", userId).Return(10.00, true, nil)

	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitReachedShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Достигнут месячный лимит (-12.00 RUB)",
		userId, mainMenu)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	processor.EXPECT().AddExpense(ctx, 125.50, "RUB", "Кофе", date, userId)
	processor.EXPECT().GetFreeLimit(ctx, "Кофе", "RUB", userId).Return(-12.00, true, nil)

	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command: addExpenseCommand,
		UserID:  userId,
	})

	assert.NoError(t, err)
}

func TestOnGetWeekExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Недельный бюджет:\nпусто\n", userId, mainMenu)

	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	reporter.EXPECT().GetReport(ctx, model.Week, "RUB", userId).Return(&expense_reporter.ExpenseReport{IsEmpty: true}, nil)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command: getExpensesCommand,
		UserID:  userId,
	})

	assert.NoError(t, err)
}

func TestOnGetMonthExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Месячный бюджет:\nпусто\n", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)

	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	reporter.EXPECT().GetReport(ctx, model.Month, "RUB", userId).Return(&expense_reporter.ExpenseReport{IsEmpty: true}, nil)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "month",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnGetYearExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Годовой бюджет:\nпусто\n", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)

	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)
	reporter.EXPECT().GetReport(ctx, model.Year, "RUB", userId).Return(&expense_reporter.ExpenseReport{IsEmpty: true}, nil)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

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
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "invalid",
		UserID:           userId,
	})

	assert.NoError(t, err)
}

func TestOnSetLimitShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	userId := int64(100)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Установлен месячный лимит 12500.50 RUB для категории Дом", userId, mainMenu)

	processor := exp_processor_mock.NewMockExpenseProcessor(ctrl)
	reporter := exp_reporter_mock.NewMockExpenseReporter(ctrl)

	processor.EXPECT().SetLimit(ctx, "Дом", userId, 12500.50, "RUB").Return(12500.50, nil)

	model := New(sender, currencies, processor, reporter, &testLogger{}, nil, nil)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "Дом;12500.50",
		UserID:           userId,
	})

	assert.NoError(t, err)
}
