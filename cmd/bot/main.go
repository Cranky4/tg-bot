package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

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

	var wg sync.WaitGroup

	tgClient, err := tg.New(config)
	if err != nil {
		log.Fatal("tg client init failed:", err)
	}

	converter := converter.ExchConverter{}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := converter.Load(ctx); err != nil {
			log.Println("exchange load err:", err)
			return
		}
		log.Println("loaded")
	}()

	msgModel := messages.New(tgClient, storage.NewMemoryStorage(), &converter)

	tgClient.ListenUpdates(ctx, msgModel)

	wg.Wait()

	log.Println("bye...")
}
