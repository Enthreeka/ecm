package callback

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Enthreeka/tg-question-bot/internal/handler/tgbot"
	service "github.com/Enthreeka/tg-question-bot/internal/usecase"
	customErr "github.com/Enthreeka/tg-question-bot/pkg/bot_error"
	"github.com/Enthreeka/tg-question-bot/pkg/excel"
	store "github.com/Enthreeka/tg-question-bot/pkg/local_storage"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	"github.com/Enthreeka/tg-question-bot/pkg/postgres"
	customMsg "github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api"
	"github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api/markup"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type CallbackUser interface {
	AdminRoleSetting() tgbot.ViewFunc
	AdminLookUp() tgbot.ViewFunc
	AdminDeleteRole() tgbot.ViewFunc
	AdminSetRole() tgbot.ViewFunc
	MainMenu() tgbot.ViewFunc
	QuestionSettings() tgbot.ViewFunc
}

type callbackUser struct {
	userService service.UserService
	log         *logger.Logger
	store       store.LocalStorage
	tgMsg       customMsg.Message
	pg          *postgres.Postgres
	excel       *excel.Excel

	mu sync.RWMutex
}

func NewCallbackUser(
	userService service.UserService,
	log *logger.Logger,
	store store.LocalStorage,
	tgMsg customMsg.Message,
	pg *postgres.Postgres,
	excel *excel.Excel,
) (CallbackUser, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if log == nil {
		return nil, errors.New("logger is nil")
	}
	if userService == nil {
		return nil, errors.New("userService is nil")
	}
	if tgMsg == nil {
		return nil, errors.New("tgMsg is nil")
	}
	if pg == nil {
		return nil, errors.New("pg is nil")
	}
	if excel == nil {
		return nil, errors.New("excel is nil")
	}

	return &callbackUser{
		userService: userService,
		log:         log,
		store:       store,
		tgMsg:       tgMsg,
		pg:          pg,
		excel:       excel,
	}, nil
}

// AdminRoleSetting -  user_setting
func (c *callbackUser) AdminRoleSetting() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Управление администраторами"

		if _, err := c.tgMsg.SendEditMessage(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			&markup.UserSetting,
			text); err != nil {
			return err
		}

		return nil
	}
}

// AdminLookUp - admin_look_up
func (c *callbackUser) AdminLookUp() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		admin, err := c.userService.GetAllAdmin(ctx)
		if err != nil {
			c.log.Error("AdminLookUp: UserRepo.GetAllAdmin: %v", err)
			return customErr.ErrServerError
		}

		adminByte, err := json.MarshalIndent(admin, "", "\t")
		if err != nil {
			c.log.Error("AdminLookUp: json.MarshalIndent: %v", err)
			return customErr.ErrServerError
		}

		if _, err := c.tgMsg.SendEditMessage(update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			&markup.MainMenu,
			string(adminByte)); err != nil {
			return err
		}

		return nil
	}
}

// AdminDeleteRole - admin_delete_role
func (c *callbackUser) AdminDeleteRole() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Напишите никнейм пользователя, у которого вы хотите отозвать права администратором.\nДля отмены команды" +
			"отправьте /cancel"

		msgID, err := c.tgMsg.SendNewMessage(update.CallbackQuery.Message.Chat.ID, nil, text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			OperationType: store.AdminDelete,
			CurrentMsgID:  msgID,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
		}, update.CallbackQuery.Message.Chat.ID)

		return nil
	}
}

// AdminSetRole - admin_set_role
func (c *callbackUser) AdminSetRole() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		text := "Напишите никнейм пользователя, которого вы хотите назначить администратором.\nДля отмены команды " +
			"отправьте /cancel"

		msgID, err := c.tgMsg.SendNewMessage(update.CallbackQuery.Message.Chat.ID, nil, text)
		if err != nil {
			return err
		}

		c.store.Set(&store.Data{
			OperationType: store.AdminCreate,
			CurrentMsgID:  msgID,
			PreferMsgID:   update.CallbackQuery.Message.MessageID,
		}, update.CallbackQuery.Message.Chat.ID)

		return nil
	}
}

func (c *callbackUser) MainMenu() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {

		if _, err := c.tgMsg.SendEditMessage(update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			&markup.StartMenu,
			"Панель управления"); err != nil {
			return err
		}

		return nil
	}
}

func (c *callbackUser) QuestionSettings() tgbot.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		rows, err := c.pg.Pool.Query(ctx, `SELECT id,user_id,question from question `)
		if err != nil {
			return err
		}
		defer rows.Close()

		var results []excel.Question
		for rows.Next() {
			var result excel.Question
			err := rows.Scan(&result.ID, &result.UserID, &result.Question)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		c.mu.Lock()
		fileName, err := c.excel.GenerateUserResultsExcelFile(results, update.CallbackQuery.From.UserName)
		if err != nil {
			c.log.Error("Excel.GenerateExcelFile: failed to generate excel file: %v", err)
			return err
		}

		fileIDBytes, err := c.excel.GetExcelFile(fileName)
		if err != nil {
			c.log.Error("Excel.GetExcelFile: failed to get excel file: %v", err)
			return err
		}
		c.mu.Unlock()

		if fileIDBytes == nil {
			c.log.Error("fileIDBytes is nil")
			return errors.New("ошибка в обработке файла")
		}

		if _, err := c.tgMsg.SendDocument(update.FromChat().ID,
			fileName,
			fileIDBytes,
			"Список вопрососов",
		); err != nil {
			return err
		}

		return nil
	}
}
