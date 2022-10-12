package expenses_memory_repo

import (
	"context"
	"strings"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
)

type repository struct {
	expenses []*model.Expense
	limits   map[string]int64
}

func NewRepository() repo.ExpensesRepository {
	return &repository{
		limits: make(map[string]int64),
	}
}

func (r *repository) Add(ctx context.Context, ex model.Expense) error {
	r.expenses = append(r.expenses, &ex)

	return nil
}

func (r *repository) GetExpenses(ctx context.Context, p model.ExpensePeriod) ([]*model.Expense, error) {
	exps := make([]*model.Expense, 0, len(r.expenses))

	periodStart := p.GetStart(time.Now())

	for i := 0; i < len(r.expenses); i++ {
		if r.expenses[i].Datetime.After(periodStart) {
			exps = append(exps, r.expenses[i])
		}
	}

	return exps, nil
}

func (r *repository) SetLimit(ctx context.Context, category string, amount int64) error {
	r.limits[strings.ToLower(category)] = amount

	return nil
}

func (r *repository) GetFreeLimit(ctx context.Context, category string) (int64, bool, error) {
	loweredCategory := strings.ToLower(category)
	if _, ex := r.limits[loweredCategory]; !ex {
		return 0, false, nil
	}

	year, month, _ := time.Now().Date()
	loc := time.Now().Location()
	beginingOfMonth := time.Date(year, month, 0, 0, 0, 0, 0, loc)

	var total int64
	for i := 0; i < len(r.expenses); i++ {
		if strings.ToLower(r.expenses[i].Category) == loweredCategory &&
			r.expenses[i].Datetime.After(beginingOfMonth) {
			total += r.expenses[i].Amount
		}
	}

	return r.limits[loweredCategory] - total, true, nil
}
