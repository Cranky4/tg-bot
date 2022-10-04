package converter

import (
	"context"
	"encoding/json"
	"net/http"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	CNY Currency = "CNY"
	RUB Currency = "RUB"
)

var AvailableCurrencies = []Currency{USD, EUR, CNY, RUB}

type Converter interface {
	Load(ctx context.Context) error
	FromRUB(amount float64, to Currency) float64
	ToRUB(amount float64, from Currency) float64
}

// https://api.exchangerate.host/latest?base=RUB&symbols=USD,CNY,EUR
type ExchConverter struct {
	Rates *Rates
}

type Rates struct {
	CNY float64
	EUR float64
	USD float64
}

type ExchangeResponse struct {
	Success bool
	Base    string
	Rates   Rates
}

func (c *ExchConverter) FromRUB(amount float64, to Currency) float64 {
	switch to {
	case USD:
		return c.Rates.USD * amount
	case CNY:
		return c.Rates.CNY * amount
	case EUR:
		return c.Rates.EUR * amount
	case RUB:
		return amount
	default:
		return amount
	}
}

func (c *ExchConverter) ToRUB(amount float64, from Currency) float64 {
	switch from {
	case USD:
		return amount / c.Rates.USD
	case CNY:
		return amount / c.Rates.CNY
	case EUR:
		return amount / c.Rates.EUR
	case RUB:
		return amount
	default:
		return amount
	}
}

func (c *ExchConverter) Load(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://api.exchangerate.host/latest?base=RUB&symbols=USD,CNY,EUR",
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result ExchangeResponse
	json.NewDecoder(resp.Body).Decode(&result)

	c.Rates = &result.Rates

	return nil
}
