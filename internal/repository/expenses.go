package repository

import "gitlab.ozon.dev/cranky4/tg-bot/internal/utils/expenses"

type ExpensesRepository interface {
	Add(expense expenses.Expense) error
	GetExpenses(period expenses.ExpensePeriod) ([]*expenses.Expense, error)
	SetLimit(category string, amount int64) error
	GetFreeLimit(category string) (int64, bool, error)
}
