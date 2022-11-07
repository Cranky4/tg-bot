package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker/kafka"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	memoryrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/memory"
	sqlrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/sql"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache"
	memory_cache "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/memory"
	redis_cache "gitlab.ozon.dev/cranky4/tg-bot/internal/service/cache/redis"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/expense_reporter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	reportrequestreceiver "gitlab.ozon.dev/cranky4/tg-bot/internal/service/report_request_receiver"
)

const undefinedMode = "неизвестный режим кеширования: %s"

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
	converter := serviceconverter.NewConverter(exchangerate.NewGetter())
	cache, err := initCache(*config)
	if err != nil {
		log.Fatal(err.Error())
	}
	expenseReporter := expense_reporter.NewReporter(repo, converter, cache)

	reportReceiver := reportrequestreceiver.NewReportRequestReceiver(broker, config.MessageBroker.Queue, expenseReporter)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logger.Info("Start receiving reports requests")

	if err := reportReceiver.Start(ctx); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Stop receiving reports requests")
}

func initCache(conf config.Config) (cache.Cache, error) {
	switch conf.Cache.Mode {
	case cache.MemoryMode:
		return memory_cache.NewLRUCache(conf.Cache.Length), nil
	case cache.RedisMode:
		return redis_cache.NewRedisCache(conf.Redis), nil
	default:
		return nil, fmt.Errorf(undefinedMode, conf.Cache.Mode)
	}
}

func initMessageBroker(conf config.MessageBrokerConf) (messagebroker.MessageBroker, error) {
	switch conf.Adapter {
	case "kafka":
		return kafka.NewKafkaCient(conf)
	}

	return nil, errors.New("Невалидный адаптер брокера сообщений")
}

func initRepo(conf config.Config) repo.ExpensesRepository {
	var repo repo.ExpensesRepository
	var err error

	switch conf.Storage.Mode {
	case "memory":
		repo = memoryrepo.NewRepository()
	case "sql":
		repo, err = sqlrepo.NewRepository(conf.Database)
		if err != nil {
			log.Fatalf("cannot connect to db %s", err.Error())
		}
	default:
		log.Fatalf("unknown repo mode %s", conf.Storage)
	}

	return repo
}
