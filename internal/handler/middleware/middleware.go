package middleware

import (
	"context"
	"errors"
	"github.com/Enthreeka/tg-question-bot/internal/handler/tgbot"
	service "github.com/Enthreeka/tg-question-bot/internal/usecase"
	customErr "github.com/Enthreeka/tg-question-bot/pkg/bot_error"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AdminMiddleware(service service.UserService, next tgbot.ViewFunc) tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		user, err := service.GetUserByID(ctx, update.FromChat().ID)
		if err != nil {
			if errors.Is(err, customErr.ErrNoRows) {
				return nil
			}
			return err
		}

		if user.UserRole == "admin" || user.UserRole == "superAdmin" {
			return next(ctx, bot, update)
		}

		// пользователь не админ
		return nil
	}
}
