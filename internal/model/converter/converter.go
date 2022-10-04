package converter

import (
	"context"
	"math"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	CNY Currency = "CNY"
	RUB Currency = "RUB"

	precisionFactor = 10000 // конвертация валют идут с точностью до 0.0001
)

type Converter interface {
	Load(ctx context.Context) error
	FromRUB(amount float64, to Currency) float64
	ToRUB(amount float64, from Currency) float64
	GetAvailableCurrencies() map[Currency]struct{}
}

type exchConverter struct {
	rates  exchangerate.Rates
	getter exchangerate.RatesGetter
}

func NewConverter(getter exchangerate.RatesGetter) Converter {
	return &exchConverter{
		getter: getter,
	}
}

func (c *exchConverter) FromRUB(amount float64, to Currency) float64 {
	var multiplier float64

	switch to {
	case USD:
		multiplier = c.rates.USD
	case CNY:
		multiplier = c.rates.CNY
	case EUR:
		multiplier = c.rates.EUR
	case RUB:
		multiplier = 1.0
	}

	return math.Round(amount*multiplier*precisionFactor) / precisionFactor
}

func (c *exchConverter) ToRUB(amount float64, from Currency) float64 {
	var divizor float64

	switch from {
	case USD:
		divizor = c.rates.USD
	case CNY:
		divizor = c.rates.CNY
	case EUR:
		divizor = c.rates.EUR
	case RUB:
		divizor = 1.0
	}

	return math.Round(amount/divizor*precisionFactor) / precisionFactor
}

func (c *exchConverter) Load(ctx context.Context) error {
	rates, err := c.getter.Get(ctx)
	if err != nil {
		return err
	}

	c.rates = rates

	return nil
}

func (c *exchConverter) GetAvailableCurrencies() map[Currency]struct{} {
	curencies := make(map[Currency]struct{})
	curencies[USD] = struct{}{}
	curencies[EUR] = struct{}{}
	curencies[CNY] = struct{}{}

	return curencies
}
