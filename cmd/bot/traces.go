package main

import (
	"io"

	jaeger_config "github.com/uber/jaeger-client-go/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

var tracesFlusher io.Closer

func initTraces() {
	cfg := jaeger_config.Configuration{
		Sampler: &jaeger_config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	var err error

	tracesFlusher, err = cfg.InitGlobalTracer("tg_bot")
	if err != nil {
		logger.Fatal("Cannot init tracing", logger.LogDataItem{Key: "error", Value: err.Error()})
	}

	logger.Debug("Трейсы готовы")
}

func flushTraces() error {
	return tracesFlusher.Close()
}
