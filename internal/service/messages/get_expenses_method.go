package servicemessages

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

const (
	reportRequestedMsg = "Запрос на формирование отчета отправлен"
)

func (m *Model) getExpenses(ctx context.Context, msg Message) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getExpenses")
	defer span.Finish()

	var expPeriod model.ExpensePeriod

	switch msg.CommandArguments {
	case "week":
		expPeriod = model.Week
	case "month":
		expPeriod = model.Month
	case "year":
		expPeriod = model.Year
	default:
		if msg.CommandArguments != "" {
			return "", errors.New(errGetExpensesInvalidPeriodMessage)
		}
		expPeriod = model.Week
	}

	err := m.reportRequester.SendRequestReport(ctx, msg.UserID, expPeriod, m.currency)
	if err != nil {
		return "", err
	}

	return reportRequestedMsg, nil
}
