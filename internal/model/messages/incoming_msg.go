package messages

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

type MessageSender interface {
	SendMessage(text string, userID int64) error
}

type Storage interface {
	Add(expenses.Expense) error
	GetExpenses(expenses.ExpensePeriod) []expenses.Expense
}

type Model struct {
	tgClient MessageSender
	storage  Storage
}

func New(tgClient MessageSender, storage Storage) *Model {
	return &Model{
		tgClient: tgClient,
		storage:  storage,
	}
}

type Message struct {
	Text   string
	UserID int64
}

type Command int64

const (
	Start Command = iota
	AddExpense
	GetExpenses
)

func (c Command) Route() string {
	switch c {
	case Start:
		return "/start"
	case AddExpense:
		return "/add-expense"
	case GetExpenses:
		return "/get-expenses"
	}

	return ""
}

func (s *Model) IncomingMessage(msg Message) error {
	parts := strings.SplitN(msg.Text, " ", 2)
	command := parts[0]

	var data string
	if len(parts) > 1 {
		data = parts[1]
	}

	switch command {
	case Start.Route():
		return s.tgClient.SendMessage("hello", msg.UserID)
	case AddExpense.Route():
		return s.addExpense(data, msg)
	case GetExpenses.Route():
		return s.getExpenses(data, msg)
	}

	return s.tgClient.SendMessage("не знаю эту команду", msg.UserID)
}

func (s *Model) addExpense(data string, msg Message) error {
	parts := strings.Split(data, ";")

	if len(parts) != 3 {
		return s.tgClient.SendMessage(
			"Неверное количество параметров.\nОжидается: Сумма;Категория;Дата \nНапример: 120.50;Дом;2022-10-01 13:25:23",
			msg.UserID,
		)
	}

	trimmedAmount := strings.Trim(parts[0], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return s.tgClient.SendMessage(
			fmt.Sprintf("Неверное значение суммы: %v", trimmedAmount), msg.UserID)
	}

	timmedDatetime := strings.Trim(parts[2], " ")
	datetime, err := time.Parse("2006-01-02 15:04:05", timmedDatetime)
	if err != nil {
		return s.tgClient.SendMessage(
			fmt.Sprintf("Неверный формат даты и времени: %v. Ожидается 2022-01-28 15:10:11", timmedDatetime), msg.UserID)
	}

	trimmedCategory := strings.Trim(parts[1], " ")
	err = s.storage.Add(expenses.Expense{
		Amount:   int(float32(amount) * 100),
		Category: trimmedCategory,
		Datetime: datetime,
	})
	if err != nil {
		return s.tgClient.SendMessage("Ошибка сохранения траты", msg.UserID)
	}

	return s.tgClient.SendMessage(
		fmt.Sprintf("Трата %.02fр добавлена в категорию %s с датой %s", amount, trimmedCategory, timmedDatetime),
		msg.UserID,
	)
}

func (s *Model) getExpenses(period string, msg Message) error {
	var expPeriod expenses.ExpensePeriod

	switch period {
	case "week":
		expPeriod = expenses.Week
	case "month":
		expPeriod = expenses.Month
	case "year":
		expPeriod = expenses.Year
	default:
		if period != "" {
			return s.tgClient.SendMessage(
				"Неверный период. Ожидается: year, month, week. По-умолчанию week",
				msg.UserID,
			)
		}
		expPeriod = expenses.Week
	}

	expenses := s.storage.GetExpenses(expPeriod)

	result := make(map[string]int)

	for _, e := range expenses {
		_, ok := result[e.Category]
		if !ok {
			result[e.Category] = 0
		}

		result[e.Category] += e.Amount
	}

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", expPeriod),
	)
	defer reporter.Reset()

	if len(result) == 0 {
		reporter.WriteString("пусто\n")
	}

	for cat, amount := range result {
		reporter.WriteString(
			fmt.Sprintf("%s - %.02fр\n", cat, float32(amount)/100),
		)
	}

	return s.tgClient.SendMessage(
		reporter.String(),
		msg.UserID,
	)
}
