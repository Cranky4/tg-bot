package messagebroker

import "context"

type MetaItem struct {
	Key   string
	Value []byte
}

type MessageBroker interface {
	Produce(ctx context.Context, topic, key string, value []byte, meta []MetaItem) error
}
