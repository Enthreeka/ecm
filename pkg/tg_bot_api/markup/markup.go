package markup

import (
	"github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	StartMenu = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Скачать вопросы", "bot_setting")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Управление пользователями", "user_setting")),
	)

	UserSetting = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назначить роль администратора", "admin_set_role"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отозвать роль администратора", "admin_delete_role"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Посмотреть список администраторов", "admin_look_up"),
		),
		tgbotapi.NewInlineKeyboardRow(button.MainMenuButton),
	)

	SuperAdminSetting = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назначить администратором", "create_admin")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Назначить супер администратором", "create_super_admin")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Забрать права администратора", "delete_admin")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Список администраторов", "all_admin")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вернуться назад", "user_setting")),
		tgbotapi.NewInlineKeyboardRow(button.MainMenuButton),
	)

	MainMenu = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button.MainMenuButton))
)
