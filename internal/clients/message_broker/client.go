package messagebroker

import "context"

type MetaItem struct {
	Key   string
	Value any
}

type Message struct {
	Key   string
	Value []byte
	Meta  []MetaItem
}

type MessageBroker interface {
	Produce(ctx context.Context, topic string, message Message) error
	Consume(ctx context.Context, topic string, out chan<- Message) error
}
