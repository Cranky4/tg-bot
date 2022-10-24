package main

import (
	jaeger_config "github.com/uber/jaeger-client-go/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicelogger "gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

func initTraces(logger logger.Logger) {
	cfg := jaeger_config.Configuration{
		Sampler: &jaeger_config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	flusher, err := cfg.InitGlobalTracer("tg_bot")
	if err != nil {
		logger.Fatal("Cannot init tracing", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
	}
	defer func() {
		err := flusher.Close()
		if err != nil {
			logger.Error("Error while tracer flush", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()
}
