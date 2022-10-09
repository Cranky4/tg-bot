package messages

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	msgmocks "gitlab.ozon.dev/cranky4/tg-bot/internal/mocks/messages"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

type testGetter struct{}

func (g *testGetter) Get(ctx context.Context) (exchangerate.Rates, error) {
	return exchangerate.Rates{
		USD: 2,
		EUR: 3,
		CNY: 4,
	}, nil
}

var testConverter = converter.NewConverter(&testGetter{})

func TestOnStartCommandShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := msgmocks.NewMockMessageSender(ctrl)
	storage := msgmocks.NewMockStorage(ctrl)
	model := New(sender, storage, testConverter)

	msg := "Привет, я буду считать твои деньги. Вот что я умею:\n" +
		"addExpense- добавить трату\n" +
		"Пример: /addExpense 10;Дом;2022-10-04 10:00:00\n" +
		"getExpenses - получить список трат за неделю, месяц и год\n" +
		"Пример: /getExpenses week\n" +
		"requestCurrencyChange - вызвать менюсмены валюты\n" +
		"setCurrency - установить валюту ввода и отображения отчетов.\n" +
		"Пример: /setCurrency EUR\n"

	sender.EXPECT().SendMessage(msg, int64(123), mainMenu)

	err := model.IncomingMessage(Message{
		Command: startCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnUnknownCommandShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123), mainMenu)
	storage := msgmocks.NewMockStorage(ctrl)
	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50 RUB добавлена в категорию Кофе с датой 2022-10-01 12:56:00",
		int64(123), mainMenu)

	storage := msgmocks.NewMockStorage(ctrl)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	storage.EXPECT().Add(expenses.Expense{
		Amount:   12550,
		Category: "Кофе",
		Datetime: date,
	})

	model := New(sender, storage, testConverter)

	err = model.IncomingMessage(Message{
		Command:          addExpenseCommand,
		CommandArguments: "125.50; Кофе; 2022-10-01 12:56:00",
		UserID:           123,
	})

	assert.NoError(t, err)
}

func TestOnAddExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверное количество параметров.\n"+
		"Ожидается: Сумма;Категория;Дата \n"+
		"Например: 120.50;Дом;2022-10-01 13:25:23", int64(123), mainMenu)

	storage := msgmocks.NewMockStorage(ctrl)
	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(Message{
		Command: addExpenseCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnGetExpenseShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Недельный бюджет:\nпусто\n", int64(123), mainMenu)

	storage := msgmocks.NewMockStorage(ctrl)
	storage.EXPECT().GetExpenses(expenses.Week)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(Message{
		Command: getExpensesCommand,
		UserID:  123,
	})

	assert.NoError(t, err)
}

func TestOnGetExpenseShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := msgmocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("неверный период. Ожидается: year, month, week. По-умолчанию week", int64(123), mainMenu)

	storage := msgmocks.NewMockStorage(ctrl)

	model := New(sender, storage, testConverter)

	err := model.IncomingMessage(Message{
		Command:          getExpensesCommand,
		CommandArguments: "wrong",
		UserID:           123,
	})

	assert.NoError(t, err)
}
