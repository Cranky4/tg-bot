package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/exchangerate"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/tg"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/converter"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/storage"
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

	go func(ctx context.Context) {
		if err := converter.Load(ctx); err != nil {
			log.Println("exchange load err:", err)
			return
		}
		log.Println("loaded")
	}(ctx)

	go func(ctx context.Context) {
		<-ctx.Done()

		tgClient.Stop()

		log.Println("receiving stopped...")
	}(ctx)

	msgModel := messages.New(tgClient, storage.NewMemoryStorage(), converter)

	tgClient.ListenUpdates(msgModel)

	log.Println("bye...")
}
