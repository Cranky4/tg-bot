package exchangerate

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

const URL = "https://api.exchangerate.host/latest?base=RUB&symbols=USD,CNY,EUR"

type RatesGetter interface {
	Get(ctx context.Context) (*ExchangeResponse, error)
}

type exchRatesGetter struct {
	url    string
	logger logger.Logger
}

func NewGetter(logger logger.Logger) RatesGetter {
	return &exchRatesGetter{
		url:    URL,
		logger: logger,
	}
}

type Rates struct {
	CNY float64
	EUR float64
	USD float64
}

type ExchangeResponse struct {
	Success bool
	Base    string
	Data    string // 2022-10-14
	Rates   Rates
}

func (g *exchRatesGetter) Get(ctx context.Context) (*ExchangeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
	if err != nil {
		return nil, err
	}

	g.logger.Debug("получение курсов валют", logger.LogDataItem{Key: "URL", Value: g.url})

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
	}()

	var result ExchangeResponse
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	g.logger.Debug("получение курсов валют успешно")

	return &result, err
}
