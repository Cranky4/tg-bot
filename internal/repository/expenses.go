package repository

import (
	"context"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

type ExpensesRepository interface {
	Add(ctx context.Context, expense model.Expense) error
	GetExpenses(ctx context.Context, period model.ExpensePeriod, userId int64) ([]*model.Expense, error)
	SetLimit(ctx context.Context, category string, userId, amount int64) error
	GetFreeLimit(ctx context.Context, category string, userId int64) (int64, bool, error)
}
