package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/api"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/tracer"
	pkg_api "gitlab.ozon.dev/cranky4/tg-bot/pkg/reporter_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	pkg_api.UnimplementedReporterV1Server
	messagesService *servicemessages.Model
}

func (s *server) SendReport(ctx context.Context, request *pkg_api.SendReportRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GRPCServer_GetReport")

	md, ok := metadata.FromIncomingContext(ctx)
	var err error
	if ok {
		span, ctx, err = extractTraceFromMeta(ctx, span, md)
		if err != nil {
			return nil, err
		}
	}
	defer span.Finish()

	report := expense_reporter.ExpenseReport{
		Rows:   request.GetRows(),
		UserID: request.GetUserId(),
	}

	err = s.messagesService.SendReport(ctx, &report)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func initGRPСServer(grpcConf config.GRPCConf, messagesService *servicemessages.Model) error {
	grpcPort := fmt.Sprintf(":%d", grpcConf.Port)

	grpcListener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	s := grpc.NewServer(
		grpc.InTapHandle(api.CountRequestsInterceptor),
		grpc.ChainUnaryInterceptor(api.LogInterceptor, api.TracingInterceptor),
	)
	pkg_api.RegisterReporterV1Server(s, &server{messagesService: messagesService})

	logger.Info("GRPC server listening " + grpcPort)
	if err = s.Serve(grpcListener); err != nil {
		logger.Fatal(fmt.Sprintf("failed to serve: %v", err))
	}

	return nil
}

func initHTTPServer(httpConf config.HTTPConf, grpcConf config.GRPCConf) error {
	httpPort := fmt.Sprintf(":%d", httpConf.Port)
	grpcPort := fmt.Sprintf(":%d", grpcConf.Port)

	ctx := context.Background()
	mux := runtime.NewServeMux()

	err := pkg_api.RegisterReporterV1HandlerFromEndpoint(
		ctx,
		mux,
		grpcPort,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		return err
	}

	logger.Info("HTTP server listening " + httpPort)
	// G114: Use of net/http serve function that has no support for setting timeouts
	if err = http.ListenAndServe(httpPort, mux); err != nil { //nolint:gosec
		return err
	}

	return nil
}

func extractTraceFromMeta(ctx context.Context, span opentracing.Span, md metadata.MD) (opentracing.Span, context.Context, error) {
	for k, v := range md {
		if k == "trace" {
			incomingTrace, err := tracer.ExtractTracerContext([]byte(v[0]))
			if err != nil {
				return nil, ctx, err
			}

			newSpan, newCtx := opentracing.StartSpanFromContext(ctx, "GRPCServer_GetReport", ext.RPCServerOption(incomingTrace))

			if err != nil {
				return span, ctx, err
			}

			return newSpan, newCtx, nil

		}
	}

	return span, ctx, nil
}
