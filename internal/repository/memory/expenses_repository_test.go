package expenses_memory_repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

func TestStorageShouldAddExpensesToStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewRepository()
	userId := int64(100)

	exps, err := storage.GetExpenses(ctx, model.Week, userId)
	assert.Len(t, exps, 0)
	assert.NoError(t, err)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastMonth := now.AddDate(0, -1, 1) // без 1 дня месяц назад
	lastYear := now.AddDate(-1, 0, 1)  // без 1 дня год назад

	err = storage.Add(ctx, model.Expense{
		Amount:   12000,
		Category: "Кофе",
		Datetime: now,
		UserId:   userId,
	})
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(ctx, model.Week, userId)
	assert.Len(t, exps, 1)
	assert.NoError(t, err)

	err = storage.Add(ctx, model.Expense{
		Amount:   12500,
		Category: "Еще кофе",
		Datetime: yesterday,
		UserId:   userId,
	})
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(ctx, model.Month, userId)
	assert.Len(t, exps, 2)
	assert.NoError(t, err)

	err = storage.Add(ctx, model.Expense{
		Amount:   12500,
		Category: "Еще кофе в прошлом месяце",
		Datetime: lastMonth,
		UserId:   userId,
	})
	assert.NoError(t, err)

	err = storage.Add(ctx, model.Expense{
		Amount:   12500,
		Category: "Еще кофе в прошлом году",
		Datetime: lastYear,
	})
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(ctx, model.Month, userId)
	assert.Len(t, exps, 3)
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(ctx, model.Year, userId)
	assert.Len(t, exps, 4)
	assert.NoError(t, err)

	exps, err = storage.GetExpenses(ctx, model.Week, userId)
	assert.Len(t, exps, 2)
	assert.NoError(t, err)
}

func TestStorageShouldSetLimitAndReachedIt(t *testing.T) {
	repo := NewRepository()
	now := time.Now()
	category := "Кофе"
	ctx := context.Background()
	userId := int64(100)

	err := repo.Add(ctx, model.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
		UserId:   userId,
	})
	assert.NoError(t, err)

	err = repo.Add(ctx, model.Expense{
		Amount:   25000,
		Category: "Дорогой кофе в прошлом месяце",
		Datetime: now.AddDate(0, -1, -1),
		UserId:   userId,
	})
	assert.NoError(t, err)

	freeLimit, isSet, err := repo.GetFreeLimit(ctx, category, userId)
	assert.NoError(t, err)
	assert.False(t, isSet)
	assert.Equal(t, int64(0), freeLimit)

	err = repo.SetLimit(ctx, category, userId, 25000)
	assert.NoError(t, err)

	freeLimit, isSet, err = repo.GetFreeLimit(ctx, category, userId)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, int64(13000), freeLimit)

	err = repo.Add(ctx, model.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
		UserId:   userId,
	})
	assert.NoError(t, err)

	freeLimit, isSet, err = repo.GetFreeLimit(ctx, category, userId)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, int64(1000), freeLimit)

	err = repo.Add(ctx, model.Expense{
		Amount:   12000,
		Category: category,
		Datetime: now,
		UserId:   userId,
	})
	assert.NoError(t, err)

	freeLimit, isSet, err = repo.GetFreeLimit(ctx, category, userId)
	assert.NoError(t, err)
	assert.True(t, isSet)
	assert.Equal(t, int64(-11000), freeLimit)
}
