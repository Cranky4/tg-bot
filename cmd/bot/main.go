package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/tg"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	memoryrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/memory"
	sqlrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/sql"
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

	var repo repo.ExpensesRepository
	switch config.Storage.Mode {
	case "memory":
		repo = memoryrepo.NewRepository()
	case "sql":
		repo, err = sqlrepo.NewRepository(config.Database)
		if err != nil {
			log.Fatalf("cannot connect to db %s", err.Error())
		}
	default:
		log.Fatalf("unknown repo mode %s", config.Storage)
	}

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

	// Запускаем http-сервер для метрик
	go func() {
		http.Handle(config.Metrics.URL, promhttp.Handler())

		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Metrics.Port), nil); err != nil {
			logger.Error("ошибка старта сервера метрик", servicelogger.LogDataItem{Key: "error", Value: err.Error()})
		}
	}()

	// Метрики
	labelNames := []string{"command"}
	requestsTotalCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Total count of requests",
			Name:      "requests_total",
		},
		labelNames,
	)
	responseTimeSummary := promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "tg_bot",
			Subsystem: "tg_client",
			Help:      "Response time of commands",
			Name:      "response_time_milliseconds",
			Objectives: map[float64]float64{
				0.5:  50,
				0.9:  10,
				0.99: 1,
			},
		}, labelNames,
	)

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
