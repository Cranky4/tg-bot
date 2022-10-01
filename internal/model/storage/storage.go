package storage

import (
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

type Storage interface {
	Add(expenses.Expense) error
	GetExpenses(expenses.ExpensePeriod) []expenses.Expense
}

type MemoryStorage struct {
	expenses []expenses.Expense
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		expenses: make([]expenses.Expense, 0),
	}
}

func (s *MemoryStorage) Add(ex expenses.Expense) error {
	s.expenses = append(s.expenses, ex)

	return nil
}

func (s *MemoryStorage) GetExpenses(p expenses.ExpensePeriod) []expenses.Expense {
	exps := make([]expenses.Expense, 0, len(s.expenses))

	periodStart := p.GetStart(time.Now())

	for _, ex := range s.expenses {
		if ex.Datetime.After(periodStart) {
			exps = append(exps, ex)
		}
	}

	return exps
}
