package servicemessages

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

func (m *Model) getExpenses(ctx context.Context, msg Message) (string, error) {
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

	report, err := m.expenseReporter.GetReport(ctx, expPeriod, m.currency, msg.UserID)
	if err != nil {
		return "", err
	}

	var reporter strings.Builder
	reporter.WriteString(
		fmt.Sprintf("%s бюджет:\n", &expPeriod),
	)
	defer reporter.Reset()

	if report.IsEmpty {
		reporter.WriteString("пусто\n")
	}

	for category, amount := range report.Rows {
		if _, err := reporter.WriteString(fmt.Sprintf("%s - %.02f %s\n", category, amount, m.currency)); err != nil {
			return "", err
		}
	}

	return reporter.String(), nil
}
