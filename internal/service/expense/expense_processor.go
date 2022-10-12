package expense_service

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
)

const (
	primitiveCurrencyMultiplier = 100

	errSaveExpenseMessage = "ошибка сохранения траты"
	errSetLimitMessage    = "ошибка создания лимита"
)

type ExpenseProcessor interface {
	AddExpense(ctx context.Context, amount float64, currency string, category string, datetime time.Time) (*model.Expense, error)
	GetFreeLimit(ctx context.Context, category, currency string) (float64, bool, error)
	SetLimit(ctx context.Context, category string, amount float64, currency string) (float64, error)
}

type processor struct {
	repo      repo.ExpensesRepository
	converter serviceconverter.Converter
}

func NewProcessor(repo repo.ExpensesRepository, conv serviceconverter.Converter) ExpenseProcessor {
	return &processor{
		repo:      repo,
		converter: conv,
	}
}

func (p *processor) AddExpense(ctx context.Context, amount float64, currency string, category string, datetime time.Time) (*model.Expense, error) {
	convertedAmount := p.converter.ToRUB(amount, currency)

	ex := model.Expense{
		Amount:   int64(convertedAmount * primitiveCurrencyMultiplier),
		Category: strings.Trim(category, " "),
		Datetime: datetime,
	}

	if err := p.repo.Add(ctx, ex); err != nil {
		return nil, errors.Wrap(err, errSaveExpenseMessage)
	}

	return &ex, nil
}

func (p *processor) GetFreeLimit(ctx context.Context, category, currency string) (float64, bool, error) {
	freeLimit, hasLimit, err := p.repo.GetFreeLimit(ctx, strings.Trim(category, " "))
	if err != nil {
		return 0, false, errors.Wrap(err, errSaveExpenseMessage)
	}

	convertedFreeLimit := p.converter.FromRUB(float64(freeLimit), currency)

	return convertedFreeLimit / primitiveCurrencyMultiplier, hasLimit, nil
}

func (p *processor) SetLimit(ctx context.Context, category string, amount float64, currency string) (float64, error) {
	convertedAmount := p.converter.ToRUB(amount, currency)

	if err := p.repo.SetLimit(ctx, category, int64(convertedAmount*primitiveCurrencyMultiplier)); err != nil {
		return 0, errors.Wrap(err, errSetLimitMessage)
	}

	return convertedAmount, nil
}
