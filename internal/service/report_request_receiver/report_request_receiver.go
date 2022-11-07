package reportrequestreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/api"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequester "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReportRequestReceiver interface {
	Start(ctx context.Context) error
}

type reportRequestReceiver struct {
	broker          messagebroker.MessageBroker
	queue           string
	expenseReporter expense_reporter.ExpenseReporter
	grpcConfig      config.GRPCConf
}

func NewReportRequestReceiver(
	broker messagebroker.MessageBroker,
	queue string,
	expenseReporter expense_reporter.ExpenseReporter,
	grpcConfig config.GRPCConf,
) ReportRequestReceiver {
	return &reportRequestReceiver{
		broker:          broker,
		queue:           queue,
		expenseReporter: expenseReporter,
		grpcConfig:      grpcConfig,
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

			logger.Debug(fmt.Sprintf("%v", report))

			if err := r.sendReport(ctx, report); err != nil {
				return err
			}

		}
	}
}

func (r *reportRequestReceiver) sendReport(ctx context.Context, report *expense_reporter.ExpenseReport) error {
	addr := fmt.Sprintf(":%d", r.grpcConfig.Port)

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	c := api.NewReporterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = c.SendReport(ctx, &api.SendReportRequest{
		IsEmpty: report.IsEmpty,
		Rows:    report.Rows,
		UserId:  report.UserID,
	})
	if err != nil {
		return err
	}

	return err
}
