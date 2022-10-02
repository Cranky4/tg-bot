package storage

import (
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
)

type Storage interface {
	Add(expense expenses.Expense) error
	GetExpenses(period expenses.ExpensePeriod) []*expenses.Expense
}

type memoryStorage struct {
	expenses []expenses.Expense
}

func NewMemoryStorage() Storage {
	return &memoryStorage{}
}

func (m *memoryStorage) Add(ex expenses.Expense) error {
	m.expenses = append(m.expenses, ex)

	return nil
}

func (m *memoryStorage) GetExpenses(p expenses.ExpensePeriod) []*expenses.Expense {
	exps := make([]*expenses.Expense, 0, len(m.expenses))

	periodStart := p.GetStart(time.Now())

	for i := 0; i < len(m.expenses); i++ {
		if m.expenses[i].Datetime.After(periodStart) {
			exps = append(exps, &m.expenses[i])
		}
	}

	return exps
}
