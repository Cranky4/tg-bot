package expense_processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
)

const (
	primitiveCurrencyMultiplier = 100

	errSaveExpenseMessage = "ошибка сохранения траты"
	errSetLimitMessage    = "ошибка создания лимита"
)

type ExpenseProcessor interface {
	AddExpense(ctx context.Context, amount float64, currency string, category string, datetime time.Time, userId int64) (*model.Expense, error)
	GetFreeLimit(ctx context.Context, category, currency string, userId int64) (float64, bool, error)
	SetLimit(ctx context.Context, category string, userId int64, amount float64, currency string) (float64, error)
}

type processor struct {
	repo      repo.ExpensesRepository
	converter serviceconverter.Converter
	cache     cache.Cache
}

func NewProcessor(repo repo.ExpensesRepository, conv serviceconverter.Converter, cache cache.Cache) ExpenseProcessor {
	return &processor{
		repo:      repo,
		converter: conv,
		cache:     cache,
	}
}

func (p *processor) AddExpense(ctx context.Context, amount float64, currency string, category string, datetime time.Time, userId int64) (*model.Expense, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AddExpense")
	defer span.Finish()

	convertedAmount := p.converter.ToRUB(amount, currency)

	ex := model.Expense{
		Amount:   int64(convertedAmount * primitiveCurrencyMultiplier),
		Category: strings.Trim(category, " "),
		Datetime: datetime,
		UserId:   userId,
	}

	if err := p.repo.Add(ctx, ex); err != nil {
		return nil, errors.Wrap(err, errSaveExpenseMessage)
	}

	// сбрасываем кеш при добавлении новой траты
	for _, period := range []model.ExpensePeriod{model.Week, model.Month, model.Year} {
		cacheKey := fmt.Sprintf("%d-%v-%s", userId, period, time.Now().Format("2006-01-02"))
		_, err := p.cache.Del(ctx, cacheKey)
		if err != nil {
			return nil, err
		}
	}

	return &ex, nil
}

func (p *processor) GetFreeLimit(ctx context.Context, category, currency string, userId int64) (float64, bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetFreeLimit")
	defer span.Finish()

	freeLimit, hasLimit, err := p.repo.GetFreeLimit(ctx, strings.Trim(category, " "), userId)
	if err != nil {
		return 0, false, errors.Wrap(err, errSaveExpenseMessage)
	}

	convertedFreeLimit := p.converter.FromRUB(float64(freeLimit), currency)

	return convertedFreeLimit / primitiveCurrencyMultiplier, hasLimit, nil
}

func (p *processor) SetLimit(ctx context.Context, category string, userId int64, amount float64, currency string) (float64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SetLimit")
	defer span.Finish()

	convertedAmount := p.converter.ToRUB(amount, currency)

	if err := p.repo.SetLimit(ctx, category, userId, int64(convertedAmount*primitiveCurrencyMultiplier)); err != nil {
		return 0, errors.Wrap(err, errSetLimitMessage)
	}

	return convertedAmount, nil
}
