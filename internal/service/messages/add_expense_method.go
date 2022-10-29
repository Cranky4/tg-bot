package servicemessages

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

func (m *Model) addExpense(ctx context.Context, msg Message) (string, error) {
	// Трейсы
	span, ctx := opentracing.StartSpanFromContext(ctx, "addExpense")
	defer span.Finish()

	parts := strings.Split(msg.CommandArguments, ";")

	if len(parts) != 3 {
		return "", errors.New(errAddExpenseInvalidParameterMessage)
	}

	trimmedAmount := strings.Trim(parts[0], " ")
	amount, err := strconv.ParseFloat(trimmedAmount, 32)
	if err != nil {
		return "", fmt.Errorf(errInvalidAmountParameterMessage, trimmedAmount)
	}

	trimmedDatetime := strings.Trim(parts[2], " ")
	datetime, err := time.Parse(datetimeFormat, trimmedDatetime)
	if err != nil {
		return "", fmt.Errorf(errAddExpenseInvalidDatetimeParameterMessage, trimmedDatetime)
	}

	trimmedCategory := strings.Trim(parts[1], " ")

	if _, err = m.expenseProcessor.AddExpense(ctx, amount, m.currency, trimmedCategory, datetime, msg.UserID); err != nil {
		return "", err
	}

	freeLimit, hasLimit, err := m.expenseProcessor.GetFreeLimit(ctx, trimmedCategory, m.currency, msg.UserID)
	if err != nil {
		return "", err
	}

	var responseMsg string
	responseMsg = msgExpenseAdded

	if hasLimit {
		var addMsg string

		if freeLimit > 0 {
			addMsg = msgFreeLimit
		} else {
			addMsg = msgLimitReached
		}
		responseMsg = fmt.Sprintf(
			"%s.\n%s", responseMsg,
			fmt.Sprintf(addMsg, freeLimit, m.currency),
		)
	}

	return fmt.Sprintf(responseMsg, amount, m.currency, trimmedCategory, trimmedDatetime), nil
}
