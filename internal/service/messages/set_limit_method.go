package servicemessages

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (m *Model) setLimit(ctx context.Context, msg Message) (string, error) {
	parts := strings.Split(msg.CommandArguments, ";")

	if len(parts) != 2 {
		return "", errors.New(errSetLimitInvalidParameterMessage)
	}

	trimmedCategory := strings.Trim(parts[0], " ")

	trimmedAmount := strings.Trim(parts[1], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return "", fmt.Errorf(errInvalidAmountParameterMessage, trimmedAmount)
	}

	convertedAmount, err := m.expenseProcessor.SetLimit(ctx, trimmedCategory, msg.UserID, amount, m.currency)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(msgSetLimit, convertedAmount, m.currency, trimmedCategory), nil
}
