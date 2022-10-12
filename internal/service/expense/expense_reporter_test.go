package expense_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repomocks "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/mocks"
)

func TestGetReportWithSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	repo := repomocks.NewMockExpensesRepository(ctrl)
	date, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 12:56:00")
	assert.NoError(t, err)

	reporter := NewReporter(repo, testConverter)

	repo.EXPECT().GetExpenses(ctx, model.Week).Return([]*model.Expense{
		{
			Amount:   12550,
			Category: "Категория",
			Datetime: date,
		},
	}, nil)

	report, err := reporter.GetReport(ctx, model.Week, "RUB")
	assert.NoError(t, err)
	assert.False(t, report.IsEmpty)
	assert.Len(t, report.Rows, 1)
}

func TestGetReportWithEmptyReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	repo := repomocks.NewMockExpensesRepository(ctrl)

	reporter := NewReporter(repo, testConverter)

	repo.EXPECT().GetExpenses(ctx, model.Week).Return([]*model.Expense{}, nil)

	report, err := reporter.GetReport(ctx, model.Week, "RUB")
	assert.NoError(t, err)
	assert.True(t, report.IsEmpty)
	assert.Len(t, report.Rows, 0)
}

func TestGetReportWithDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	repo := repomocks.NewMockExpensesRepository(ctrl)

	reporter := NewReporter(repo, testConverter)

	repo.EXPECT().GetExpenses(ctx, model.Week).Return([]*model.Expense{}, errors.New("database error"))

	report, err := reporter.GetReport(ctx, model.Week, "RUB")
	assert.Error(t, err)
	assert.Nil(t, report)
}
