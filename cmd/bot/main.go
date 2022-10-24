package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/tg"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_processor"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	servicelogger "gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	logger, err := servicelogger.NewLogger(config.Logger.Level, config.Env)
	if err != nil {
		log.Fatal("logger init failed:", err)
	}

	tgClient, err := tg.New(config, logger)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}

	converter := serviceconverter.NewConverter(exchangerate.NewGetter(logger))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	repo := initRepo(*config)

	// Загружаем курс валют
	go func(ctx context.Context) {
		if err := converter.Load(ctx); err != nil {
			logger.Error("exchange load err", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
			return
		}
	}(ctx)

	// Выключаем слежение за обновлениями в клиенте телеги
	go func(ctx context.Context) {
		<-ctx.Done()

		tgClient.Stop()
		logger.Debug("receiving stopped...")
	}(ctx)

	// Метрики
	requestsTotalCounter := initTotalCounter()
	responseTimeSummary := initResponseTime()
	go func() {
		err := startHTTPServer(config.Metrics.URL, config.Metrics.Port)
		if err != nil {
			logger.Error("Error while tracer flush", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	// Трейсы
	initTraces(logger)
	defer func() {
		if err := flushTraces(); err != nil {
			logger.Error("traces flush err", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	messagesService := servicemessages.New(
		tgClient,
		converter.GetAvailableCurrencies(),
		expense_processor.NewProcessor(repo, converter),
		expense_reporter.NewReporter(repo, converter),
		logger,
		requestsTotalCounter,
		responseTimeSummary,
	)

	tgClient.ListenUpdates(ctx, messagesService)

	logger.Debug("bye...")
}
