package reportsender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	api "gitlab.ozon.dev/cranky4/tg-bot/pkg/reporter_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ReportSender interface {
	Send(ctx context.Context, report *expense_reporter.ExpenseReport) error
}

type reportSender struct {
	grpcConfig config.GRPCConf
}

func NewReportSender(conf config.GRPCConf) ReportSender {
	return &reportSender{
		grpcConfig: conf,
	}
}

func (s *reportSender) Send(ctx context.Context, report *expense_reporter.ExpenseReport) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ReportSender_Send")
	defer span.Finish()

	addr := fmt.Sprintf(":%d", s.grpcConfig.Port)

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	c := api.NewReporterV1Client(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	ctx = metadata.AppendToOutgoingContext(ctx, "trace", string(encodedTraceContext))

	_, err = c.SendReport(ctx, &api.SendReportRequest{
		Rows:   report.Rows,
		UserId: report.UserID,
	})

	return err
}
