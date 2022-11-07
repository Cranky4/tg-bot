package reportrequester

import (
	"context"
	"fmt"

	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

type ReportRequester interface {
	SendRequestReport(ctx context.Context, userID int64, period model.ExpensePeriod, currency string) error
}

type reportRequester struct {
	broker    messagebroker.MessageBroker
	queueName string
}

func NewReportRequester(broker messagebroker.MessageBroker, queueName string) ReportRequester {
	return &reportRequester{
		broker:    broker,
		queueName: queueName,
	}
}

func (r *reportRequester) SendRequestReport(ctx context.Context, userID int64, period model.ExpensePeriod, currency string) error {
	UID := fmt.Sprintf("%d", userID)

	return r.broker.Produce(
		ctx,
		r.queueName,
		UID,
		[]byte(fmt.Sprintf("%d", period)),
		[]messagebroker.MetaItem{
			{Key: "userId", Value: []byte(UID)},
			{Key: "currency", Value: []byte(currency)},
		},
	)
}
