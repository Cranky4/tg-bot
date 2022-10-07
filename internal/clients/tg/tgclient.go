package tg

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/model/messages"
)

type TokenGetter interface {
	Token() string
}

type Client struct {
	tgclient *tgbotapi.BotAPI
}

func New(tokenGetter TokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenGetter.Token())
	if err != nil {
		return nil, errors.Wrap(err, "NewBotAPI")
	}

	return &Client{tgclient: client}, nil
}

func (c *Client) SendMessage(text string, userID int64, buttons []string) error {
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

	if _, err := c.tgclient.Send(msg); err != nil {
		return errors.Wrap(err, "client.Send")
	}
	return nil
}

func (c *Client) ListenUpdates(msgModel *messages.Model) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates := c.tgclient.GetUpdatesChan(u)

	log.Println("listening for messages")

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			err := msgModel.IncomingMessage(messages.Message{
				Text:             update.Message.Text,
				UserID:           update.Message.From.ID,
				Command:          update.Message.Command(),
				CommandArguments: update.Message.CommandArguments(),
			})
			if err != nil {
				log.Println("error processing message:", err)
			}
		}
	}
}

func (c *Client) Stop() {
	c.tgclient.StopReceivingUpdates()
}
