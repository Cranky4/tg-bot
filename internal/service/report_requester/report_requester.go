package reportrequester

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
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
	broker                      messagebroker.MessageBroker
	queueName                   string
	totalMessageProducedCounter *prometheus.CounterVec
}

func NewReportRequester(broker messagebroker.MessageBroker, queueName string, totalMessageProducedCounter *prometheus.CounterVec) ReportRequester {
	return &reportRequester{
		broker:                      broker,
		queueName:                   queueName,
		totalMessageProducedCounter: totalMessageProducedCounter,
	}
}

func (r *reportRequester) SendRequestReport(ctx context.Context, userID int64, period model.ExpensePeriod, currency string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SendRequestReport")
	ext.SpanKindRPCClient.Set(span)
	defer span.Finish()

	UID := fmt.Sprintf("%d", userID)

	request := ReportRequest{UserID: userID, Currency: currency, Period: period}
	value, err := json.Marshal(request)
	if err != nil {
		return err
	}

	traceContext := make(map[string]string)
	err = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(traceContext),
	)
	if err != nil {
		return err
	}

	encodedTraceContext, err := json.Marshal(traceContext)
	if err != nil {
		return err
	}

	err = r.broker.Produce(
		ctx,
		r.queueName,
		messagebroker.Message{
			Key:   UID,
			Value: value,
			Meta: []messagebroker.MetaItem{
				{
					Key:   "trace",
					Value: encodedTraceContext,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	if r.totalMessageProducedCounter != nil {
		r.totalMessageProducedCounter.WithLabelValues(r.queueName).Inc()
	}

	return nil
}
