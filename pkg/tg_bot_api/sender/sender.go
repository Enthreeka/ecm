package sender

import (
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender interface {
	SendMsgToNewUser(userID int64) error
	GetSuccessCounter() int64
}

type sender struct {
	log *logger.Logger
	bot *tgbotapi.BotAPI

	successCounter int64
}

func NewSender(log *logger.Logger, bot *tgbotapi.BotAPI) *sender {
	return &sender{
		log: log,
		bot: bot,
	}
}

func (s *sender) GetSuccessCounter() int64 {
	return s.successCounter
}
