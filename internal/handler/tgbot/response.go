package tgbot

import (
	store "github.com/Enthreeka/tg-question-bot/pkg/local_storage"
	"github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	success = "Операция выполнена успешно. "
)

func (b *Bot) response(operationType store.TypeCommand, currentMessageId int, preferMessageId int, update *tgbotapi.Update) {
	var (
		messageId int
		userID    = update.FromChat().ID
	)

	if update.Message != nil {
		messageId = update.Message.MessageID
	} else if update.CallbackQuery != nil {
		messageId = update.CallbackQuery.Message.MessageID
	}

	if resp, err := b.bot.Request(tgbotapi.NewDeleteMessage(userID, messageId)); nil != err || !resp.Ok {
		b.log.Error("failed to delete message id %d (%s): %v", currentMessageId, string(resp.Result), err)
	}

	if resp, err := b.bot.Request(tgbotapi.NewDeleteMessage(userID, preferMessageId)); nil != err || !resp.Ok {
		b.log.Error("failed to delete message id %d (%s): %v", preferMessageId, string(resp.Result), err)
	}

	text, markup := responseText(operationType)
	if _, err := b.tgMsg.SendEditMessage(userID, currentMessageId, markup, text); err != nil {
		b.log.Error("failed to send telegram message: ", err)
	}
}

func responseText(operationType store.TypeCommand) (string, *tgbotapi.InlineKeyboardMarkup) {
	switch operationType {
	case store.AdminCreate:
		return success + "Пользователь получил администраторские права.", &markup.UserSetting
	case store.AdminDelete:
		return success + "Пользователь лишился администраторских прав.", &markup.UserSetting
	}
	return success, nil
}
