package main

import (
	"context"
	"fmt"
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
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicelogger "gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/metrics"
	reportrequester "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_requester"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}
	logger.SetLevel(config.Logger.Level)

	tgClient, err := tg.New(config)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}

	converter := serviceconverter.NewConverter(exchangerate.NewGetter())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	repo := initRepo(*config)

	// Загружаем курс валют
	go func(ctx context.Context) {
		if err = converter.Load(ctx); err != nil {
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
	go func() {
		err = startMetricsHTTPServer(config.Metrics.URL, config.Metrics.Port)
		if err != nil {
			logger.Error("Error while tracer flush", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	// Трейсы
	initTraces()
	defer func() {
		if err = flushTraces(); err != nil {
			logger.Error("traces flush err", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	// Кэш
	cache, err := initCache(*config)
	if err != nil {
		logger.Fatal(fmt.Sprintf("cache init failed: %s", err))
	}

	// Брокер сообщений
	broker, err := initMessageBroker(config.MessageBroker)
	if err != nil {
		logger.Fatal(fmt.Sprintf("broker message init failed: %s", err))
	}

	messagesService := servicemessages.New(
		tgClient,
		converter.GetAvailableCurrencies(),
		expense_processor.NewProcessor(repo, converter, cache),
		reportrequester.NewReportRequester(broker, config.MessageBroker.Queue, metrics.MessageBrokerMessagesProducesTotalCounter),
		metrics.TotalRequestCounter,
		metrics.ResponseTimeSummary,
	)

	// GRPC
	go func() {
		if err := initGRPСServer(config.GRPC, messagesService); err != nil {
			logger.Fatal(fmt.Sprintf("GRPC server err %s", err))
		}
	}()

	// HTTP
	go func() {
		if err := initHTTPServer(config.HTTP, config.GRPC); err != nil {
			logger.Fatal(fmt.Sprintf("HTTP server err %s", err))
		}
	}()

	tgClient.ListenUpdates(ctx, messagesService)

	logger.Debug("bye...")
}
