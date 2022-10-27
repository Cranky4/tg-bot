package servicemessages

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

func (m *Model) setCurrency(ctx context.Context, msg Message) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "setCurrency")
	defer span.Finish()

	if _, found := m.currencies[msg.CommandArguments]; !found {
		return "", fmt.Errorf(errUnknownCurrency, msg.CommandArguments)
	}

	m.currency = msg.CommandArguments

	return fmt.Sprintf(msgCurrencySet, msg.CommandArguments), nil
}
