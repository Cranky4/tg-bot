package reportrequester

import (
	"context"
	"encoding/json"
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

	value, err := json.Marshal(ReportRequest{
		Period:   model.Week,
		UserID:   123,
		Currency: "RUB",
	})

	assert.Nil(t, err)

	client.EXPECT().Produce(
		wrapedCtx,
		"queue",
		messagebroker.Message{
			Key:   "123",
			Value: value,
			Meta:  nil,
		})

	requester := NewReportRequester(client, "queue", nil)

	err = requester.SendRequestReport(ctx, 123, model.Week, "RUB")
	assert.Nil(t, err)
}
