package main

import (
	"errors"

	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker/kafka"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
)

func initMessageBroker(conf config.MessageBrokerConf) (messagebroker.MessageBroker, error) {
	switch conf.Adapter {
	case "kafka":
		return kafka.NewKafkaCient(conf)
	}

	return nil, errors.New("Невалидный адаптер брокера сообщений")
}
