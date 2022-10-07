package exchangerate

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const URL = "https://api.exchangerate.host/latest?base=RUB&symbols=USD,CNY,EUR"

type Rates struct {
	CNY float64
	EUR float64
	USD float64
}

type RatesGetter interface {
	Get(ctx context.Context) (Rates, error)
}

type exchRatesGetter struct {
	url string
}

func NewGetter() RatesGetter {
	return &exchRatesGetter{
		url: URL,
	}
}

type exchangeResponse struct {
	Rates Rates
}

func (g *exchRatesGetter) Get(ctx context.Context) (Rates, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
	if err != nil {
		return Rates{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Rates{}, err
	}
	defer res.Body.Close()

	var result exchangeResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return Rates{}, err
	}

	return result.Rates, nil
}
