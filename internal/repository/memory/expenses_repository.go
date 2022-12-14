package expenses_memory_repo

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
)

type limit struct {
	userId int64
	amount int64
}

type repository struct {
	expenses []*model.Expense
	limits   map[string]*limit
}

func NewRepository() repo.ExpensesRepository {
	return &repository{
		limits: make(map[string]*limit),
	}
}

func (r *repository) Add(ctx context.Context, ex model.Expense) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Add")
	defer span.Finish()

	r.expenses = append(r.expenses, &ex)

	return nil
}

func (r *repository) GetExpenses(ctx context.Context, p model.ExpensePeriod, userId int64) ([]*model.Expense, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetExpenses")
	defer span.Finish()

	exps := make([]*model.Expense, 0, len(r.expenses))

	periodStart := p.GetStart(time.Now())

	for i := 0; i < len(r.expenses); i++ {
		if r.expenses[i].Datetime.After(periodStart) {
			exps = append(exps, r.expenses[i])
		}
	}

	return exps, nil
}

func (r *repository) SetLimit(ctx context.Context, category string, userId, amount int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "SetLimit")
	defer span.Finish()

	r.limits[strings.ToLower(category)] = &limit{amount: amount, userId: userId}

	return nil
}

func (r *repository) GetFreeLimit(ctx context.Context, category string, userId int64) (int64, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetFreeLimit")
	defer span.Finish()

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

	return r.limits[loweredCategory].amount - total, true, nil
}
