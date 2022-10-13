package servicemessages

import "fmt"

func (m *Model) setCurrency(msg Message) (string, error) {
	if _, found := m.currencies[msg.CommandArguments]; !found {
		return "", fmt.Errorf(errUnknownCurrency, msg.CommandArguments)
	}

	m.currency = msg.CommandArguments

	return fmt.Sprintf(msgCurrencySet, msg.CommandArguments), nil
}
