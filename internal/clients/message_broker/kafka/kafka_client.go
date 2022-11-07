package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

type kafkaClient struct {
	producer sarama.SyncProducer
}

func NewKafkaCient(conf config.MessageBrokerConf) (messagebroker.MessageBroker, error) {
	producer, err := newProducer(conf)
	if err != nil {
		return nil, err
	}

	return &kafkaClient{producer: producer}, nil
}

func (c *kafkaClient) Produce(ctx context.Context, topic, key string, value []byte, meta []messagebroker.MetaItem) error {
	headers := make([]sarama.RecordHeader, 0, len(meta))
	for i := 0; i < len(meta); i++ {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(meta[i].Key),
			Value: meta[i].Value,
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: headers,
	}

	partition, offset, err := c.producer.SendMessage(msg)

	logger.Debug(
		fmt.Sprintf("[KAFKA] добавлено сообщение %v (partiiton %d, offset %d)", msg, partition, offset),
	)

	if err != nil {
		return err
	}

	return nil
}

func newProducer(conf config.MessageBrokerConf) (sarama.SyncProducer, error) {
	ver, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Version = ver

	producer, err := sarama.NewSyncProducer([]string{conf.Addr}, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}
