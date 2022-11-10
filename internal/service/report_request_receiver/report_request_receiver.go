package reportrequestreceiver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber/jaeger-client-go"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequester "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester"
	reportsender "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_sender"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/tracer"
)

type ReportRequestReceiver interface {
	Start(ctx context.Context) error
}

type reportRequestReceiver struct {
	broker                      messagebroker.MessageBroker
	queue                       string
	expenseReporter             expense_reporter.ExpenseReporter
	reportSender                reportsender.ReportSender
	totalMessageConsumedCounter *prometheus.CounterVec
}

func NewReportRequestReceiver(
	broker messagebroker.MessageBroker,
	queue string,
	expenseReporter expense_reporter.ExpenseReporter,
	reportSender reportsender.ReportSender,
	totalMessageConsumedCounter *prometheus.CounterVec,
) ReportRequestReceiver {
	return &reportRequestReceiver{
		broker:                      broker,
		queue:                       queue,
		expenseReporter:             expenseReporter,
		reportSender:                reportSender,
		totalMessageConsumedCounter: totalMessageConsumedCounter,
	}
}

func (r *reportRequestReceiver) Start(ctx context.Context) error {
	out := make(chan messagebroker.Message)
	defer close(out)

	go func() {
		err := r.broker.Consume(ctx, r.queue, out)

		if err != nil {
			logger.Error(err.Error(), logger.LogDataItem{
				Key: "service", Value: "REPORT_REQUEST_RECEIVER",
			})
		} else {
			logger.Info("done consuming", logger.LogDataItem{
				Key: "service", Value: "REPORT_REQUEST_RECEIVER",
			})
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-out:
			reportRequest := &reportrequester.ReportRequest{}

			if r.totalMessageConsumedCounter != nil {
				r.totalMessageConsumedCounter.WithLabelValues(r.queue).Inc()
			}

			span, wrapedCtx := opentracing.StartSpanFromContext(ctx, "ReportRequestReceiver")
			span, wrapedCtx, err := extractTraceFromMeta(wrapedCtx, span, msg.Meta)
			if err != nil {
				return err
			}
			defer span.Finish()

			// Меняет ид трейса для логов
			if spanCtx, ok := span.Context().(jaeger.SpanContext); ok {
				logger.SetTraceId(spanCtx.TraceID().String())
			}

			logger.Debug(fmt.Sprintf("получено сообщение %v", msg))

			err = json.Unmarshal(msg.Value, reportRequest)
			if err != nil {
				return err
			}

			report, err := r.expenseReporter.GetReport(
				wrapedCtx,
				reportRequest.Period,
				reportRequest.Currency,
				reportRequest.UserID,
			)
			if err != nil {
				return err
			}

			if err := r.reportSender.Send(wrapedCtx, report); err != nil {
				return err
			}

		}
	}
}

func extractTraceFromMeta(ctx context.Context, span opentracing.Span, meta []messagebroker.MetaItem) (opentracing.Span, context.Context, error) {
	for _, v := range meta {
		if v.Key == "trace" {
			incomingTrace, err := tracer.ExtractTracerContext(v.Value)
			if err != nil {
				return nil, ctx, err
			}

			newSpan, newCtx := opentracing.StartSpanFromContext(ctx, "ReportRequestReceiver_Receive", ext.RPCServerOption(incomingTrace))

			if err != nil {
				return nil, newCtx, err
			}

			return newSpan, newCtx, nil
		}
	}

	return span, ctx, nil
}
