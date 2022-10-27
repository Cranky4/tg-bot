package tg

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/config"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/service/logger"
	servicemessages "gitlab.ozon.dev/cranky4/tg-bot/internal/service/messages"
)

type TgClient interface {
	SendMessage(text string, userID int64, buttons []string) error
	ListenUpdates(ctx context.Context, msgModel *servicemessages.Model)
	Stop()
}

type client struct {
	api *tgbotapi.BotAPI
}

func New(tokenGetter config.TokenGetter) (TgClient, error) {
	api, err := tgbotapi.NewBotAPI(tokenGetter.GetToken())
	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &client{api: api}, nil
}

func (c *client) SendMessage(text string, userID int64, buttons []string) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	if buttons != nil {
		btns := make([][]tgbotapi.KeyboardButton, 0, len(buttons))
		for _, b := range buttons {
			btns = append(btns, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(b)})
		}

		keyboard := tgbotapi.NewReplyKeyboard(btns...)
		msg.ReplyMarkup = keyboard
	}

	if _, err := c.api.Send(msg); err != nil {
		return errors.Wrap(err, "client.Send")
	}
	return nil
}

func (c *client) ListenUpdates(ctx context.Context, msgModel *servicemessages.Model) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates := c.api.GetUpdatesChan(u)

	logger.Info("listening for messages")

	for update := range updates {
		if update.Message != nil {
			err := msgModel.IncomingMessage(ctx, servicemessages.Message{
				Text:             update.Message.Text,
				UserID:           update.Message.From.ID,
				Command:          update.Message.Command(),
				CommandArguments: update.Message.CommandArguments(),
			})
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

func (c *client) Stop() {
	c.api.StopReceivingUpdates()
}
