package converter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
)

type testRatesGetter struct{}

func (g *testRatesGetter) Get(ctx context.Context) (exchangerate.Rates, error) {
	return exchangerate.Rates{
		CNY: 2,
		USD: 3,
		EUR: 4,
	}, nil
}

type testErrorRatesGetter struct{}

func (g *testErrorRatesGetter) Get(ctx context.Context) (exchangerate.Rates, error) {
	return exchangerate.Rates{}, errors.New("timout")
}

func TestConverterShouldCorrectConvertToRUB(t *testing.T) {
	converter := NewConverter(&testRatesGetter{})
	err := converter.Load(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 100.00, converter.ToRUB(100, RUB))
	assert.Equal(t, 100.00, converter.ToRUB(200, CNY))
	assert.Equal(t, 100.00, converter.ToRUB(300, USD))
	assert.Equal(t, 100.00, converter.ToRUB(400, EUR))
}

func TestConverterShouldCorrectConvertFromRUB(t *testing.T) {
	converter := NewConverter(&testRatesGetter{})
	err := converter.Load(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 100.00, converter.FromRUB(100, RUB))
	assert.Equal(t, 200.00, converter.FromRUB(100, CNY))
	assert.Equal(t, 300.00, converter.FromRUB(100, USD))
	assert.Equal(t, 400.00, converter.FromRUB(100, EUR))
}

func TestConverterLoadError(t *testing.T) {
	converter := NewConverter(&testErrorRatesGetter{})
	err := converter.Load(context.Background())
	assert.Error(t, err)
}

func TestConverterGetAvailableCurrencies(t *testing.T) {
	curencies := make(map[string]struct{})
	curencies[USD] = struct{}{}
	curencies[EUR] = struct{}{}
	curencies[CNY] = struct{}{}
	curencies[RUB] = struct{}{}

	converter := NewConverter(&testRatesGetter{})
	curs := converter.GetAvailableCurrencies()
	assert.Equal(t, curencies, curs)
}
