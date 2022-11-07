package reportrequester

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	messagebroker "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker"
	clientmocks "gitlab.ozon.dev/cranky4/tg-bot/internal/clients/message_broker/mocks"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model"
)

func TestSendRequestReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	_, wrapedCtx := opentracing.StartSpanFromContext(ctx, "wrap1")

	client := clientmocks.NewMockMessageBroker(ctrl)
	client.EXPECT().Produce(
		wrapedCtx,
		"queue",
		messagebroker.Message{
			Key:   "123",
			Value: []byte(fmt.Sprintf("%d", model.Week)),
			Meta: []messagebroker.MetaItem{
				{Key: "userId", Value: []byte(fmt.Sprintf("%d", 123))},
				{Key: "currency", Value: []byte("RUB")},
			},
		})

	requester := NewReportRequester(client, "queue")

	err := requester.SendRequestReport(ctx, 123, model.Week, "RUB")
	assert.Nil(t, err)
}
