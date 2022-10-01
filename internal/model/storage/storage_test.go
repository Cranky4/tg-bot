package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

func Test_Storage_ShouldAddExpensesToStorage(t *testing.T) {
	storage := NewMemoryStorage()

	exps := storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 0)

	datetime, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 11:25:32")
	assert.NoError(t, err)

	err = storage.Add(expenses.Expense{
		Amount:   12000,
		Category: "Кофе",
		Datetime: datetime,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 1)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе",
		Datetime: datetime,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 2)

	datetime, err = time.Parse("2006-01-02 15:04:05", "2022-09-04 11:25:32")
	assert.NoError(t, err)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе в прошлом месяце",
		Datetime: datetime,
	})
	assert.NoError(t, err)

	exps = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 3)

	exps = storage.GetExpenses(expenses.Year)
	assert.Len(t, exps, 3)

	exps = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 2)
}
