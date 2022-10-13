package servicemessages

import (
	"sort"
	"strings"
)

func (m *Model) requestCurrencyChange() (string, []string) {
	currencies := make([]string, 0, len(m.currencies))
	for c := range m.currencies {
		currencies = append(currencies, strings.Join([]string{"/", setCurrencyCommand, " ", c}, ""))
	}

	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i] < currencies[j]
	})

	return "Выберите валюту", currencies
}
