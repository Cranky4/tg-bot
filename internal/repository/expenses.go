package repository

import (
	"context"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/expenses"
)

type ExpensesRepository interface {
	Add(ctx context.Context, expense expenses.Expense) error
	GetExpenses(ctx context.Context, period expenses.ExpensePeriod) ([]*expenses.Expense, error)
	SetLimit(ctx context.Context, category string, amount int64) error
	GetFreeLimit(ctx context.Context, category string) (int64, bool, error)
}
