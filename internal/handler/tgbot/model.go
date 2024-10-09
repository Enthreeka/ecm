package tgbot

import (
	"github.com/Enthreeka/tg-question-bot/internal/entity"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

func userUpdateToModel(update *tgbotapi.Update) *entity.User {
	user := new(entity.User)

	if update != nil {
		user.ID = update.Message.From.ID
		user.TGUsername = update.Message.From.UserName
		user.CreatedAt = time.Now().Local()
		user.UserRole = entity.UserType
	}

	return user
}
