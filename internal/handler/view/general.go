package view

import (
	"context"
	"github.com/Enthreeka/tg-question-bot/internal/handler/tgbot"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	"github.com/Enthreeka/tg-question-bot/pkg/postgres"
	customMsg "github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api"
	"github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ViewGeneral struct {
	log   *logger.Logger
	tgMsg customMsg.Message
	pg    *postgres.Postgres
}

func NewViewGeneral(
	log *logger.Logger,
	tgMsg customMsg.Message,
	pg *postgres.Postgres,
) *ViewGeneral {
	return &ViewGeneral{
		log:   log,
		tgMsg: tgMsg,
		pg:    pg,
	}
}

func (c *ViewGeneral) CallbackStartAdminPanel() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		if _, err := c.tgMsg.SendNewMessage(update.FromChat().ID, &markup.StartMenu, "Панель управления"); err != nil {
			return err
		}

		return nil
	}
}

func (c *ViewGeneral) CallbackStartUser() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		startMenu := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Перейти в канал", "https://t.me/MoscowEcon")),
		)
		if _, err := c.tgMsg.SendNewMessage(update.FromChat().ID, &startMenu, "Привет!\nЗадайте вопросы нашим аналитикам. На самые интересные из них мы ответим в Telegram-канале «Экономика Москвы»."); err != nil {
			c.log.Error("Failed to send start menu: ", err)
			return nil
		}

		return nil
	}
}
