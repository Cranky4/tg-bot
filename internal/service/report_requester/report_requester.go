package reportrequester

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opentracing/opentracing-go"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

type ReportRequest struct {
	UserID   int64
	Period   model.ExpensePeriod
	Currency string
}

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
	span, ctx := opentracing.StartSpanFromContext(ctx, "SendRequestReport")
	defer span.Finish()

	UID := fmt.Sprintf("%d", userID)

	request := ReportRequest{UserID: userID, Currency: currency, Period: period}
	value, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return r.broker.Produce(
		ctx,
		r.queueName,
		messagebroker.Message{
			Key:   UID,
			Value: value,
		},
	)
}
