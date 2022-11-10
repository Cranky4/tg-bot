package api

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"
)

func LogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	logger.Info(
		fmt.Sprintf("получен запрос %s, данные %v", info.FullMethod, req),
		logger.LogDataItem{Key: "service", Value: "GRPC Server"},
	)

	m, err := handler(ctx, req)
	return m, err
}

func CountRequestsInterceptor(ctx context.Context, info *tap.Info) (context.Context, error) {
	metrics.GRPCReqeustTotalCounter.WithLabelValues(info.FullMethodName).Inc()

	return ctx, nil
}

func MetricInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GRPC_"+info.FullMethod)
	defer span.Finish()

	m, err := handler(ctx, req)
	return m, err
}
