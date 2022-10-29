package expense_processor

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repomocks "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/mocks"
	cachemocks "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/mocks"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
)

type testGetter struct{}

func (g *testGetter) Get(ctx context.Context) (*exchangerate.ExchangeResponse, error) {
	return &exchangerate.ExchangeResponse{
		Rates: exchangerate.Rates{
			USD: 2,
			EUR: 3,
			CNY: 4,
		},
	}, nil
}

var testConverter = serviceconverter.NewConverter(&testGetter{})

func TestAddExpenseWillReturnExpense(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)
	cache.EXPECT().Del(fmt.Sprintf("%d-%v-%s", userId, model.Week, time.Now().Format("2006-01-02")))
	cache.EXPECT().Del(fmt.Sprintf("%d-%v-%s", userId, model.Month, time.Now().Format("2006-01-02")))
	cache.EXPECT().Del(fmt.Sprintf("%d-%v-%s", userId, model.Year, time.Now().Format("2006-01-02")))

	processor := NewProcessor(repo, testConverter, cache)

	repo.EXPECT().Add(wrapedCtx, model.Expense{
		Amount:   12550,
		Category: "Категория",
		Datetime: date,
		UserId:   userId,
	})

	exp, err := processor.AddExpense(ctx, 125.50, "RUB", "Категория", date, userId)
	assert.NotNil(t, exp.ID)
	assert.NoError(t, err)
}

func TestAddExpenseWillReturnRepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)
	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	repo.EXPECT().Add(wrapedCtx, model.Expense{
		Amount:   12550,
		Category: "Категория",
		Datetime: date,
		UserId:   userId,
	}).Return(errors.New("database error"))

	exp, err := processor.AddExpense(ctx, 125.50, "RUB", "Категория", date, userId)
	assert.Nil(t, exp)
	assert.Error(t, err)
}

func TestGetFreeLimitWithSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	repo.EXPECT().GetFreeLimit(wrapedCtx, "Категория", userId).Return(int64(10000), true, nil)

	limit, has, err := processor.GetFreeLimit(ctx, "Категория", "RUB", userId)
	assert.Equal(t, 100.00, limit)
	assert.True(t, has)
	assert.NoError(t, err)
}

func TestGetFreeLimitWithNoLimitSet(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	repo.EXPECT().GetFreeLimit(wrapedCtx, "Категория", userId).Return(int64(0), false, nil)

	limit, has, err := processor.GetFreeLimit(ctx, "Категория", "RUB", userId)
	assert.Equal(t, 0.0, limit)
	assert.False(t, has)
	assert.NoError(t, err)
}

func TestGetFreeLimitWithRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	repo.EXPECT().GetFreeLimit(wrapedCtx, "Категория", userId).Return(int64(0), false, errors.New("database error"))

	limit, has, err := processor.GetFreeLimit(ctx, "Категория", "RUB", userId)
	assert.Equal(t, 0.0, limit)
	assert.False(t, has)
	assert.Error(t, err)
}

func TestSetLimitWithSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	repo.EXPECT().SetLimit(wrapedCtx, "Категория", userId, int64(100000))

	limit, err := processor.SetLimit(ctx, "Категория", userId, 1000.00, "RUB")
	assert.Equal(t, 1000.0, limit)
	assert.NoError(t, err)
}

func TestSetLimitWithRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)

	processor := NewProcessor(repo, testConverter, cache)

	repo.EXPECT().SetLimit(wrapedCtx, "Категория", userId, int64(100000)).Return(errors.New("database error"))

	limit, err := processor.SetLimit(ctx, "Категория", userId, 1000.00, "RUB")
	assert.Equal(t, 0.0, limit)
	assert.Error(t, err)
}
