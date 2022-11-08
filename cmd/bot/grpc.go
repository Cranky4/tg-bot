package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/api"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/tap"
	"google.golang.org/protobuf/types/known/emptypb"
)

var GRPCRequestsCountMetric = initGRPCTotalCounter()

type server struct {
	api.UnimplementedReporterServer
	messagesService *servicemessages.Model
}

func (s *server) SendReport(ctx context.Context, request *api.SendReportRequest) (*emptypb.Empty, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GRPC_SendReport")
	defer span.Finish()

	report := expense_reporter.ExpenseReport{
		IsEmpty: request.IsEmpty,
		Rows:    request.Rows,
		UserID:  request.UserId,
	}

	err := s.messagesService.SendReport(&report)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func initGRPСServer(grpcConf config.GRPCConf, httpConf config.HTTPConf, messagesService *servicemessages.Model) error {
	grpcPort := fmt.Sprintf(":%d", grpcConf.Port)

	grpcListener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	s := grpc.NewServer(
		grpc.InTapHandle(countRequestsInterceptor),
		grpc.UnaryInterceptor(logInterceptor),
	)
	api.RegisterReporterServer(s, &server{messagesService: messagesService})

	ctx := context.Background()
	rmux := runtime.NewServeMux()
	mux := http.NewServeMux()
	mux.Handle("/", rmux)

	err = api.RegisterReporterHandlerServer(ctx, rmux, &server{})
	if err != nil {
		return err
	}

	httpPort := fmt.Sprintf(":%d", httpConf.Port)
	httpListener, err := net.Listen("tcp", httpPort)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to listen http: %s", err))
	}
	runHttp := func() {
		// Register reflection service on gRPC server.
		reflection.Register(s)
		if err = s.Serve(grpcListener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}

	logger.Info("GRPC server listening " + grpcPort)
	go runHttp()

	logger.Info("HTTP server listening " + httpPort)
	// G114: Use of net/http serve function that has no support for setting timeouts
	if err = http.Serve(httpListener, mux); err != nil { //nolint:gosec
		return err
	}

	return nil
}

func logInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	logger.Info(
		fmt.Sprintf("получен запрос %s, данные %v", info.FullMethod, req),
		logger.LogDataItem{Key: "service", Value: "GRPC Server"},
	)

	m, err := handler(ctx, req)
	return m, err
}

func countRequestsInterceptor(ctx context.Context, info *tap.Info) (context.Context, error) {
	defer func(command string) {
		GRPCRequestsCountMetric.WithLabelValues(command).Inc()
	}(info.FullMethodName)

	return ctx, nil
}
