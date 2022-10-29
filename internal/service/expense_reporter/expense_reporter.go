package expense_reporter

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
)

const (
	primitiveCurrencyMultiplier = 100
	dateFormat                  = "2006-01-02 15:04:05"
)

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
	cache     cache.Cache
}

func NewReporter(repo repo.ExpensesRepository, conv serviceconverter.Converter, cache cache.Cache) ExpenseReporter {
	return &reporter{
		repo:      repo,
		converter: conv,
		cache:     cache,
	}
}

func (r *reporter) GetReport(ctx context.Context, period model.ExpensePeriod, currency string, userId int64) (*ExpenseReport, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetReport")
	defer span.Finish()

	report, ok := r.getCached(ctx, userId, period)
	if ok {
		return report, nil
	}

	expenses, err := r.repo.GetExpenses(ctx, period, userId)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64) // [категория]сумма
	report = &ExpenseReport{
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

	r.cache.Set(getCacheKey(userId, period), report)

	return report, nil
}

func (r *reporter) getCached(ctx context.Context, userId int64, period model.ExpensePeriod) (*ExpenseReport, bool) {
	span, _ := opentracing.StartSpanFromContext(ctx, "getCached")
	defer span.Finish()

	value, ok := r.cache.Get(getCacheKey(userId, period))
	if ok {
		report, ok := value.(*ExpenseReport)
		if ok {
			return report, true
		}
	}
	return nil, false
}

func getCacheKey(userId int64, period model.ExpensePeriod) string {
	return fmt.Sprintf("%d-%v-%s", userId, period, time.Now().Format("2006-01-02"))
}
