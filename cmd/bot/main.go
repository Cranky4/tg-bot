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
	repo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository"
	memoryrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/memory"
	sqlrepo "gitlab.ozon.dev/cranky4/tg-bot/internal/repository/sql"
	serviceconverter "gitlab.ozon.dev/cranky4/tg-bot/internal/service/converter"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	tgClient, err := tg.New(config)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}

	converter := serviceconverter.NewConverter(exchangerate.NewGetter())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var repo repo.ExpensesRepository
	switch config.Storage().Mode {
	case "memory":
		repo = memoryrepo.NewRepository()
	case "sql":
		repo = sqlrepo.NewRepository(config.Database())
	default:
		log.Fatalf("unknown repo mode %s", config.Storage())
	}

	// Загружаем курс валют
	go func(ctx context.Context) {
		if err := converter.Load(ctx); err != nil {
			log.Println("exchange load err:", err)
			return
		}
		log.Println("loaded")
	}(ctx)

	// Выключаем слежение за обновлениями в клиенте телеги
	go func(ctx context.Context) {
		<-ctx.Done()

		tgClient.Stop()

		log.Println("receiving stopped...")
	}(ctx)

	messagesSerbice := servicemessages.New(tgClient, repo, converter)

	tgClient.ListenUpdates(ctx, messagesSerbice)

	log.Println("bye...")
}
