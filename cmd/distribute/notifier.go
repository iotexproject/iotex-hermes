package distribute

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/iotexproject/iotex-hermes/util"
)

type Notifier struct {
	bot *tgbotapi.BotAPI
}

func NewNotifier() (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(util.MustFetchNonEmptyParam("TG_TOKEN"))
	if err != nil {
		return nil, fmt.Errorf("create telegram bot error: %v", err)
	}
	return &Notifier{bot}, nil
}

func (u *Notifier) SendMessage(message string) error {
	chatId := util.MustFetchNonEmptyParam("TG_CHATID")
	cid, _ := strconv.ParseInt(chatId, 10, 64)
	msg := tgbotapi.NewMessage(cid, message)
	_, err := u.bot.Send(msg)
	return err
}
