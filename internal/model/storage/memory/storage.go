package memorystorage

import (
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
)

type storage struct {
	expenses []*expenses.Expense
}

func NewStorage() messages.Storage {
	return &storage{}
}

func (m *storage) Add(ex expenses.Expense) error {
	m.expenses = append(m.expenses, &ex)

	return nil
}

func (m *storage) GetExpenses(p expenses.ExpensePeriod) ([]*expenses.Expense, error) {
	exps := make([]*expenses.Expense, 0, len(m.expenses))

	periodStart := p.GetStart(time.Now())

	for i := 0; i < len(m.expenses); i++ {
		if m.expenses[i].Datetime.After(periodStart) {
			exps = append(exps, m.expenses[i])
		}
	}

	return exps, nil
}
