package main

import (
	"log"

	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/tg"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
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

	msgModel := messages.New(tgClient)

	tgClient.ListenUpdates(msgModel)
}
