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
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
	memorystorage "gitlab.ozon.dev/cranky4/tg-bot/internal/model/storage/memory"
	sqlstorage "gitlab.ozon.dev/cranky4/tg-bot/internal/model/storage/sql"
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

	converter := converter.NewConverter(exchangerate.NewGetter())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	var storage messages.Storage
	switch config.Storage().Driver {
	case "memory":
		storage = memorystorage.NewStorage()
	case "sql":
		storage = sqlstorage.NewStorage(ctx, config.Database())
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

	msgModel := messages.New(tgClient, storage, converter)

	tgClient.ListenUpdates(msgModel)

	log.Println("bye...")
}
