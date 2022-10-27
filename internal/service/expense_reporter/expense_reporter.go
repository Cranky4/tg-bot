package expense_reporter

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
)

const primitiveCurrencyMultiplier = 100

type ExpenseReporter interface {
	GetReport(ctx context.Context, period model.ExpensePeriod, currencty string, userId int64) (*ExpenseReport, error)
}

type ExpenseReport struct {
	IsEmpty bool
	Rows    map[string]float64
}

type reporter struct {
	repo      repo.ExpensesRepository
	converter serviceconverter.Converter
}

func NewReporter(repo repo.ExpensesRepository, conv serviceconverter.Converter) ExpenseReporter {
	return &reporter{
		repo:      repo,
		converter: conv,
	}
}

func (r *reporter) GetReport(ctx context.Context, period model.ExpensePeriod, currency string, userId int64) (*ExpenseReport, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetReport")
	defer span.Finish()

	expenses, err := r.repo.GetExpenses(ctx, period, userId)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64) // [категория]сумма
	report := &ExpenseReport{
		Rows: make(map[string]float64),
	}

	for _, e := range expenses {
		if e.UserId == userId {
			result[e.Category] += e.Amount
		}
	}

	if len(result) == 0 {
		report.IsEmpty = true
	}

	for category, amount := range result {
		converted := r.converter.FromRUB(float64(amount/primitiveCurrencyMultiplier), currency)

		report.Rows[category] = converted
	}

	return report, nil
}
