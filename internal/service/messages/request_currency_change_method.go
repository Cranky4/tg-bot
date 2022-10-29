package servicemessages

import (
	"context"
	"sort"
	"strings"

	"github.com/opentracing/opentracing-go"
)

func (m *Model) requestCurrencyChange(ctx context.Context) (string, []string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "requestCurrencyChange")
	defer span.Finish()

	currencies := make([]string, 0, len(m.currencies))
	for c := range m.currencies {
		currencies = append(currencies, strings.Join([]string{"/", setCurrencyCommand, " ", c}, ""))
	}

	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i] < currencies[j]
	})

	return "Выберите валюту", currencies
}
