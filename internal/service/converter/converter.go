package serviceconverter

import (
	"context"
	"math"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
)

const (
	USD = "USD"
	EUR = "EUR"
	CNY = "CNY"
	RUB = "RUB"

	precisionFactor = 10000 // конвертация валют идет с точностью до 0.0001
)

type Converter interface {
	Load(ctx context.Context) error
	FromRUB(amount float64, to string) float64
	ToRUB(amount float64, from string) float64
	GetAvailableCurrencies() map[string]struct{}
}

type exchConverter struct {
	rates  Rates
	getter exchangerate.RatesGetter
}

type Rates struct {
	CNY float64
	EUR float64
	USD float64
}

func NewConverter(getter exchangerate.RatesGetter) Converter {
	return &exchConverter{
		getter: getter,
	}
}

func (c *exchConverter) FromRUB(amount float64, to string) float64 {
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

func (c *exchConverter) ToRUB(amount float64, from string) float64 {
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
	res, err := c.getter.Get(ctx)
	if err != nil {
		return err
	}

	c.rates = Rates{
		USD: res.Rates.USD,
		CNY: res.Rates.CNY,
		EUR: res.Rates.EUR,
	}

	return nil
}

func (c *exchConverter) GetAvailableCurrencies() map[string]struct{} {
	curencies := make(map[string]struct{})
	curencies[USD] = struct{}{}
	curencies[EUR] = struct{}{}
	curencies[CNY] = struct{}{}
	curencies[RUB] = struct{}{}

	return curencies
}
