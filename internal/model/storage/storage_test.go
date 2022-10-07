package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

func TestStorageShouldAddExpensesToStorage(t *testing.T) {
	storage := NewMemoryStorage()

	exps := storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 0)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastMonth := now.AddDate(0, -1, 1) // без 1 дня месяц назад
	lastYear := now.AddDate(-1, 0, 1)  // без 1 дня год назад

	err := storage.Add(expenses.Expense{
		Amount:   12000,
		Category: "Кофе",
		Datetime: now,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 1)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе",
		Datetime: yesterday,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 2)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе в прошлом месяце",
		Datetime: lastMonth,
	})
	assert.NoError(t, err)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе в прошлом году",
		Datetime: lastYear,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 3)

	exps = storage.GetExpenses(expenses.Year)
	assert.Len(t, exps, 4)

	exps = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 2)
}
