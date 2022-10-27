package expense_reporter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repomocks "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/mocks"
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
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	reporter := NewReporter(repo, testConverter)

	repo.EXPECT().GetExpenses(wrapedCtx, model.Week, userId).Return([]*model.Expense{
		{
			Amount:   12550,
			Category: "Категория",
			Datetime: date,
			UserId:   userId,
		},
	}, nil)

	report, err := reporter.GetReport(ctx, model.Week, "RUB", userId)
	assert.NoError(t, err)
	assert.False(t, report.IsEmpty)
	assert.Len(t, report.Rows, 1)
}

func TestGetReportWithEmptyReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	reporter := NewReporter(repo, testConverter)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	repo.EXPECT().GetExpenses(wrapedCtx, model.Week, userId).Return([]*model.Expense{}, nil)

	report, err := reporter.GetReport(ctx, model.Week, "RUB", userId)
	assert.NoError(t, err)
	assert.True(t, report.IsEmpty)
	assert.Len(t, report.Rows, 0)
}

func TestGetReportWithDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repomocks.NewMockExpensesRepository(ctrl)
	userId := int64(100)

	ctx := context.Background()
	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	reporter := NewReporter(repo, testConverter)

	repo.EXPECT().GetExpenses(wrapedCtx, model.Week, userId).Return([]*model.Expense{}, errors.New("database error"))

	report, err := reporter.GetReport(ctx, model.Week, "RUB", userId)
	assert.Error(t, err)
	assert.Nil(t, report)
}
