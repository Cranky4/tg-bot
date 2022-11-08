package kafka

import (
	"github.com/Shopify/sarama"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

type ConsumeHandler struct {
	out chan<- messagebroker.Message
}

func (h *ConsumeHandler) Setup(sarama.ConsumerGroupSession) error {
	logger.Debug("consumer - setup", logger.LogDataItem{Key: "service", Value: "Kafka"})
	return nil
}

func (h *ConsumeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Debug("consumer - cleanup", logger.LogDataItem{Key: "service", Value: "Kafka"})
	return nil
}

func (h *ConsumeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		session.MarkMessage(message, "")

		for message := range claim.Messages() {

			meta := make([]messagebroker.MetaItem, len(message.Headers))

			for _, header := range message.Headers {
				meta = append(meta, messagebroker.MetaItem{
					Key:   string(header.Key),
					Value: header.Value,
				})
			}

			h.out <- messagebroker.Message{
				Key:   string(message.Key),
				Value: message.Value,
				Meta:  meta,
			}
		}
	}

	return nil
}
