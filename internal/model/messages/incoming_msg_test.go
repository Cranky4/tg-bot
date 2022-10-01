package messages

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/cranky4/tg-bot/internal/mocks/messages"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockMessageSender(ctrl)
	storage := mocks.NewMockStorage(ctrl)
	model := New(sender, storage)

	sender.EXPECT().SendMessage("hello", int64(123))

	err := model.IncomingMessage(Message{
		Text:   "/start",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("не знаю эту команду", int64(123))
	storage := mocks.NewMockStorage(ctrl)
	model := New(sender, storage)

	err := model.IncomingMessage(Message{
		Text:   "some text",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnAddExpense_ShouldAnswerWithSuccessMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Трата 125.50р добавлена в категорию Кофе с датой 2022-10-01 12:56:00", int64(123))

	storage := mocks.NewMockStorage(ctrl)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	storage.EXPECT().Add(expenses.Expense{
		Amount:   12550,
		Category: "Кофе",
		Datetime: date,
	})
	model := New(sender, storage)

	err = model.IncomingMessage(Message{
		Text:   "/add-expense 125.50; Кофе; 2022-10-01 12:56:00",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnAddExpense_ShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Неверное количество параметров.\n"+
		"Ожидается: Сумма;Категория;Дата \n"+
		"Например: 120.50;Дом;2022-10-01 13:25:23", int64(123))

	storage := mocks.NewMockStorage(ctrl)
	model := New(sender, storage)

	err := model.IncomingMessage(Message{
		Text:   "/add-expense",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnGetExpense_ShouldAnswerWithEmptyMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Недельный бюджет:\nпусто\n", int64(123))

	storage := mocks.NewMockStorage(ctrl)
	storage.EXPECT().GetExpenses(expenses.Week)

	model := New(sender, storage)

	err := model.IncomingMessage(Message{
		Text:   "/get-expenses",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnGetExpense_ShouldAnswerWithFailMessage(t *testing.T) {
	ctrl := gomock.NewController(t)

	sender := mocks.NewMockMessageSender(ctrl)
	sender.EXPECT().SendMessage("Неверный период. Ожидается: year, month, week. По-умолчанию week", int64(123))

	storage := mocks.NewMockStorage(ctrl)

	model := New(sender, storage)

	err := model.IncomingMessage(Message{
		Text:   "/get-expenses wrong",
		UserID: 123,
	})

	assert.NoError(t, err)
}
