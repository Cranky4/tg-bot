package memorystorage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

func TestStorageShouldAddExpensesToStorage(t *testing.T) {
	storage := NewStorage()

	exps, err := storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 0)
	assert.NoError(t, err)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastMonth := now.AddDate(0, -1, 1) // без 1 дня месяц назад
	lastYear := now.AddDate(-1, 0, 1)  // без 1 дня год назад

	err = storage.Add(expenses.Expense{
		Amount:   12000,
		Category: "Кофе",
		Datetime: now,
	})
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 1)
	assert.NoError(t, err)

	err = storage.Add(expenses.Expense{
		Amount:   12500,
		Category: "Еще кофе",
		Datetime: yesterday,
	})
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 2)
	assert.NoError(t, err)

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

	exps, err = storage.GetExpenses(expenses.Month)
	assert.Len(t, exps, 3)
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(expenses.Year)
	assert.Len(t, exps, 4)
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(expenses.Week)
	assert.Len(t, exps, 2)
	assert.NoError(t, err)
}

func TestStorageShouldSetLimitAndReachedIt(t *testing.T) {
	storage := NewStorage()
	now := time.Now()
	category := "Кофе"

	err := storage.Add(expenses.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
	})
	assert.NoError(t, err)

	err = storage.Add(expenses.Expense{
		Amount:   25000,
		Category: "Дорогой кофе в прошлом месяце",
		Datetime: now.AddDate(0, -1, -1),
	})
	assert.NoError(t, err)

	freeLimit, isSet, err := storage.GetFreeLimit(category)
	assert.NoError(t, err)
	assert.False(t, isSet)
	assert.Equal(t, 0, freeLimit)

	err = storage.SetLimit(category, 25000)
	assert.NoError(t, err)

	freeLimit, isSet, err = storage.GetFreeLimit(category)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, 13000, freeLimit)

	err = storage.Add(expenses.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
	})
	assert.NoError(t, err)

	freeLimit, isSet, err = storage.GetFreeLimit(category)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, 1000, freeLimit)

	err = storage.Add(expenses.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
	})
	assert.NoError(t, err)

	freeLimit, isSet, err = storage.GetFreeLimit(category)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, -11000, freeLimit)
}
