package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequestreceiver "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_request_receiver"
	reportsender "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_sender"
)

const (
	startListeningInfoMsg = "слушатель запросов на формирование отчетов запущен"
	stopListeningInfoMsg  = "слушатель запросов на формирование отчетов остановлен"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}
	logger.SetLevel(config.Logger.Level)

	broker, err := initMessageBroker(config.MessageBroker)
	if err != nil {
		log.Fatal(err.Error())
	}

	repo := initRepo(*config)
	cache, err := initCache(*config)
	if err != nil {
		log.Fatal(err.Error())
	}

	converter := serviceconverter.NewConverter(exchangerate.NewGetter())
	expenseReporter := expense_reporter.NewReporter(repo, converter, cache)
	reportSender := reportsender.NewReportSender(config.GRPC)

	brokerMessageConsumedCounter := initMessageBrokerMessagesConsumedTotalCounter()
	go func() {
		err = startMetricsHTTPServer(config.Metrics.URL, config.Metrics.Port+1) // hack!
		if err != nil {
			logger.Error("Error while tracer flush", logger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	reportReceiver := reportrequestreceiver.NewReportRequestReceiver(
		broker,
		config.MessageBroker.Queue,
		expenseReporter,
		reportSender,
		brokerMessageConsumedCounter,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logger.Info(startListeningInfoMsg)

	if err := reportReceiver.Start(ctx); err != nil {
		logger.Error(err.Error())
	}

	logger.Info(stopListeningInfoMsg)
}
