package tg_bot_api

import (
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Message interface {
	SendNewMessage(chatID int64, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error)
	SendEditMessage(chatID int64, messageID int, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error)
	SendDocument(chatID int64, fileName string, fileIDBytes *[]byte, text string) (int, error)
}

type TelegramMsg struct {
	log *logger.Logger
	bot *tgbotapi.BotAPI
}

func NewMessageSetting(bot *tgbotapi.BotAPI, log *logger.Logger) *TelegramMsg {
	return &TelegramMsg{
		bot: bot,
		log: log,
	}
}

func (t *TelegramMsg) SendNewMessage(chatID int64, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if markup != nil {
		msg.ReplyMarkup = &markup
	}

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send message", zap.Error(err))
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func (t *TelegramMsg) SendEditMessage(chatID int64, messageID int, markup *tgbotapi.InlineKeyboardMarkup, text string) (int, error) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = tgbotapi.ModeHTML

	if markup != nil {
		msg.ReplyMarkup = markup
	}

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send msg: %v", err)
		return 0, err
	}

	return sendMsg.MessageID, nil
}

func (t *TelegramMsg) SendDocument(chatID int64, fileName string, fileIDBytes *[]byte, text string) (int, error) {
	msg := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: *fileIDBytes,
	})
	msg.ParseMode = tgbotapi.ModeHTML
	msg.Caption = text

	sendMsg, err := t.bot.Send(msg)
	if err != nil {
		t.log.Error("failed to send msg: %v", err)
		return 0, err
	}

	return sendMsg.MessageID, nil
}
