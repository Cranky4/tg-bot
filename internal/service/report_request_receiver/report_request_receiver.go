package reportrequestreceiver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opentracing/opentracing-go"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequester "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester"
	reportsender "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_sender"
)

type ReportRequestReceiver interface {
	Start(ctx context.Context) error
}

type reportRequestReceiver struct {
	broker          messagebroker.MessageBroker
	queue           string
	expenseReporter expense_reporter.ExpenseReporter
	reportSender    reportsender.ReportSender
}

func NewReportRequestReceiver(
	broker messagebroker.MessageBroker,
	queue string,
	expenseReporter expense_reporter.ExpenseReporter,
	reportSender reportsender.ReportSender,
) ReportRequestReceiver {
	return &reportRequestReceiver{
		broker:          broker,
		queue:           queue,
		expenseReporter: expenseReporter,
		reportSender:    reportSender,
	}
}

func (r *reportRequestReceiver) Start(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ReportRequestReceiver_Start")
	defer span.Finish()

	out := make(chan messagebroker.Message)
	defer close(out)

	go func() {
		err := r.broker.Consume(ctx, r.queue, out)
		if err != nil {
			logger.Error("[REPORT_REQUEST_RECEIVER]" + err.Error())
		} else {
			logger.Info("[REPORT_REQUEST_RECEIVER] done consuming")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-out:
			reportRequest := &reportrequester.ReportRequest{}

			logger.Debug(fmt.Sprintf("получено сообщение %v", msg))

			err := json.Unmarshal(msg.Value, reportRequest)
			if err != nil {
				return err
			}

			report, err := r.expenseReporter.GetReport(
				ctx,
				reportRequest.Period,
				reportRequest.Currency,
				reportRequest.UserID,
			)
			if err != nil {
				return err
			}

			if err := r.reportSender.Send(ctx, report); err != nil {
				return err
			}

		}
	}
}
