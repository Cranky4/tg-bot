package servicemessages

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	repomocks "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/mocks"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	msgmocks "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages/mocks"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/expenses"
)

type testGetter struct{}

func (g *testGetter) Get(ctx context.Context) (exchangerate.Rates, error) {
	return exchangerate.Rates{
		USD: 2,
		EUR: 3,
		CNY: 4,
	}, nil
}

var testConverter = serviceconverter.NewConverter(&testGetter{})

func TestOnStartCommandShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := msgmocks.NewMockMessageSender(ctrl)
	storage := repomocks.NewMockExpensesRepository(ctrl)
	model := New(sender, storage, testConverter)
	ctx := context.Background()

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

	sender.EXPECT().SendMessage(msg, int64(123), mainMenu)

	err := model.IncomingMessage(ctx, Message{
		Command: startCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnUnknownCommandShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123), mainMenu)
	storage := repomocks.NewMockExpensesRepository(ctrl)
	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00",
		int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	storage.EXPECT().Add(ctx, expenses.Expense{
		Amount:   12550,
		Category: "Кофе",
		Datetime: date,
	})
	storage.EXPECT().GetFreeLimit(ctx, "Кофе")

	model := New(sender, storage, testConverter)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitSetShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Свободный месячный лимит 10.00 RUB",
		int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	storage.EXPECT().Add(ctx, expenses.Expense{
		Amount:   12550,
		Category: "Кофе",
		Datetime: date,
	})
	storage.EXPECT().GetFreeLimit(ctx, "Кофе").Return(int64(1000), true, nil)

	model := New(sender, storage, testConverter)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseWithLimitReachedShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00.\n"+
		"Достигнут месячный лимит (-12.00 RUB)",
		int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	storage.EXPECT().Add(ctx, expenses.Expense{
		Amount:   12550,
		Category: "Кофе",
		Datetime: date,
	})
	storage.EXPECT().GetFreeLimit(ctx, "Кофе").Return(int64(-1200), true, nil)

	model := New(sender, storage, testConverter)

	err = model.IncomingMessage(ctx, Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверное количество параметров.\n"+
		"Ожидается: Сумма;Категория;Дата \n"+
		"Например: 120.50;Дом;2022-10-01 13:25:23", int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command: addExpenseCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnGetWeekExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Недельный бюджет:\nпусто\n", int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	storage.EXPECT().GetExpenses(ctx, expenses.Week)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command: getExpensesCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnGetMonthExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Месячный бюджет:\nпусто\n", int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	storage.EXPECT().GetExpenses(ctx, expenses.Month)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "month",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnGetYearExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Годовой бюджет:\nпусто\n", int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	storage.EXPECT().GetExpenses(ctx, expenses.Year)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command:          getExpensesCommand,
		CommandArguments: "year",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnGetExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверный период. Ожидается: year, month, week. По-умолчанию week", int64(123), mainMenu)

	storage := repomocks.NewMockExpensesRepository(ctrl)

	model := New(sender, storage, testConverter)

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

	storage := repomocks.NewMockExpensesRepository(ctrl)

	model := New(sender, storage, testConverter)

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

	storage := repomocks.NewMockExpensesRepository(ctrl)

	model := New(sender, storage, testConverter)

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

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage(
		"неизвестная валюта FOO",
		int64(123),
		mainMenu,
	)

	storage := repomocks.NewMockExpensesRepository(ctrl)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command:          setCurrencyCommand,
		CommandArguments: "FOO",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnSetLimitShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage(
		"неверное количество параметров.\nОжидается: Категория;Сумма \nНапример: Дом;12000.50",
		int64(123),
		mainMenu,
	)

	storage := repomocks.NewMockExpensesRepository(ctrl)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "invalid",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnSetLimitShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage(
		"Установлен месячный лимит 12500.50 RUB для категории Дом",
		int64(123),
		mainMenu,
	)

	storage := repomocks.NewMockExpensesRepository(ctrl)
	storage.EXPECT().SetLimit(ctx, "Дом", int64(1250050))

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(ctx, Message{
		Command:          setLimitCommand,
		CommandArguments: "Дом;12500.50",
		UserID:           123,
	})

	assert.NoError(t, err)
}
