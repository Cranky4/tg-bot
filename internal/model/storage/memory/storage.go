package memorystorage

import (
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/expenses"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
)

type storage struct {
	expenses []*expenses.Expense
	limits   map[string]int
}

func NewStorage() messages.Storage {
	return &storage{
		limits: make(map[string]int),
	}
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

func (m *storage) SetLimit(category string, amount int) error {
	m.limits[strings.ToLower(category)] = amount

	return nil
}

func (m *storage) GetFreeLimit(category string) (int, bool, error) {
	loweredCategory := strings.ToLower(category)
	if _, ex := m.limits[loweredCategory]; !ex {
		return 0, false, nil
	}

	year, month, _ := time.Now().Date()
	loc := time.Now().Location()
	beginingOfMonth := time.Date(year, month, 0, 0, 0, 0, 0, loc)

	total := 0
	for i := 0; i < len(m.expenses); i++ {
		if strings.ToLower(m.expenses[i].Category) == loweredCategory &&
			m.expenses[i].Datetime.After(beginingOfMonth) {
			total += m.expenses[i].Amount
		}
	}

	return m.limits[loweredCategory] - total, true, nil
}
