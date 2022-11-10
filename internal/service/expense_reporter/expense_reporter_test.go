package expense_reporter

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

func TestGetReportWithSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)
	period := model.Week

	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)
	cacheKey := fmt.Sprintf("%d-%v-%s", userId, period, time.Now().Format("2006-01-02"))
	cache.EXPECT().Get(wrapedCtx, cacheKey).Return(nil, false, nil)
	cache.EXPECT().Set(wrapedCtx, cacheKey, ExpenseReport{
		Rows:   map[string]float64{"Категория": 125},
		UserID: userId,
		Period: period,
	}, 24*time.Hour)

	reporter := NewReporter(repo, testConverter, cache)

	repo.EXPECT().GetExpenses(wrapedCtx, period, userId).Return([]*model.Expense{
		{
			Amount:   12550,
			Category: "Категория",
			Datetime: date,
			UserId:   userId,
		},
	}, nil)

	report, err := reporter.GetReport(ctx, period, "RUB", userId)
	assert.NoError(t, err)
	assert.False(t, report.IsEmpty())
	assert.Len(t, report.Rows, 1)
}

func TestGetReportWithEmptyReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)
	period := model.Week

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)
	cacheKey := fmt.Sprintf("%d-%v-%s", userId, period, time.Now().Format("2006-01-02"))
	cache.EXPECT().Get(wrapedCtx, cacheKey).Return(nil, false, nil)
	cache.EXPECT().Set(wrapedCtx, cacheKey, ExpenseReport{
		Rows:   map[string]float64{},
		UserID: userId,
		Period: period,
	}, 24*time.Hour)

	reporter := NewReporter(repo, testConverter, cache)

	repo.EXPECT().GetExpenses(wrapedCtx, period, userId).Return([]*model.Expense{}, nil)

	report, err := reporter.GetReport(ctx, period, "RUB", userId)
	assert.NoError(t, err)
	assert.True(t, report.IsEmpty())
	assert.Len(t, report.Rows, 0)
}

func TestGetReportWithDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)
	period := model.Week

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	cache := cachemocks.NewMockCache(ctrl)
	cacheKey := fmt.Sprintf("%d-%v-%s", userId, period, time.Now().Format("2006-01-02"))
	cache.EXPECT().Get(wrapedCtx, cacheKey).Return(nil, false, nil)

	reporter := NewReporter(repo, testConverter, cache)

	repo.EXPECT().GetExpenses(wrapedCtx, period, userId).Return([]*model.Expense{}, errors.New("database error"))

	report, err := reporter.GetReport(ctx, period, "RUB", userId)
	assert.Error(t, err)
	assert.Nil(t, report)
}
