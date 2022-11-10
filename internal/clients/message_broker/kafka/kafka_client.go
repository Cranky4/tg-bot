package kafka

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
)

type kafkaClient struct {
	producer      sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup
}

func NewKafkaCient(conf config.MessageBrokerConf) (messagebroker.MessageBroker, error) {
	producer, err := newProducer(conf)
	if err != nil {
		return nil, err
	}

	consumerGroup, err := newConsumerGroup(conf)
	if err != nil {
		return nil, err
	}

	return &kafkaClient{producer: producer, consumerGroup: consumerGroup}, nil
}

func (c *kafkaClient) Produce(ctx context.Context, topic string, message messagebroker.Message) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "KafkaClient_Produce")
	defer span.Finish()

	headers := make([]sarama.RecordHeader, 0, len(message.Meta))
	for i := 0; i < len(message.Meta); i++ {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(message.Meta[i].Key),
			Value: message.Meta[i].Value,
		})
	}

	logger.Debug(fmt.Sprintf("%v", message.Value))

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(message.Key),
		Value:   sarama.StringEncoder(message.Value),
		Headers: headers,
	}

	partition, offset, err := c.producer.SendMessage(msg)

	logger.Debug(
		fmt.Sprintf("добавлено сообщение %v (partiiton %d, offset %d)", msg, partition, offset),
		logger.LogDataItem{Key: "service", Value: "Kafka"},
	)

	if err != nil {
		return err
	}

	return nil
}

func (c *kafkaClient) Consume(ctx context.Context, topic string, out chan<- messagebroker.Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "KafkaClient_Consume")
	defer span.Finish()

	err := c.consumerGroup.Consume(ctx, []string{topic}, &ConsumeHandler{out: out})
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

func newConsumerGroup(conf config.MessageBrokerConf) (sarama.ConsumerGroup, error) {
	ver, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Version = ver
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup([]string{conf.Addr}, "report-request-receiver", config)
	if err != nil {
		return nil, err
	}

	return consumerGroup, nil
}
