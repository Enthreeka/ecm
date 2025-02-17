package tgbot

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Enthreeka/tg-question-bot/internal/handler"
	service "github.com/Enthreeka/tg-question-bot/internal/usecase"
	store "github.com/Enthreeka/tg-question-bot/pkg/local_storage"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	customMsg "github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"runtime/debug"
	"sync"
	"time"
)

type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error

type Bot struct {
	bot           *tgbotapi.BotAPI
	log           *logger.Logger
	store         store.LocalStorage
	tgMsg         *customMsg.TelegramMsg
	userService   service.UserService
	callbackStore *store.CallbackStorage

	cmdView      map[string]ViewFunc
	callbackView map[string]ViewFunc

	mu      sync.RWMutex
	isDebug bool
}

func NewBot(bot *tgbotapi.BotAPI,
	log *logger.Logger,
	store store.LocalStorage,
	tgMsg *customMsg.TelegramMsg,
	userService service.UserService,
	callbackStore *store.CallbackStorage,
) (*Bot, error) {
	if log == nil {
		return nil, errors.New("log is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if tgMsg == nil {
		return nil, errors.New("tgMsg is nil")
	}
	if userService == nil {
		return nil, errors.New("userService is nil")
	}
	if callbackStore == nil {
		return nil, errors.New("callbackStore is nil")
	}

	return &Bot{
		bot:           bot,
		log:           log,
		store:         store,
		tgMsg:         tgMsg,
		userService:   userService,
		callbackStore: callbackStore,
	}, nil
}

func (b *Bot) RegisterCommandView(cmd string, view ViewFunc) {
	if b.cmdView == nil {
		b.cmdView = make(map[string]ViewFunc)
	}

	b.cmdView[cmd] = view
}

func (b *Bot) RegisterCommandCallback(callback string, view ViewFunc) {
	if b.callbackView == nil {
		b.callbackView = make(map[string]ViewFunc)
	}

	b.callbackStore.AppendStorage(callback)
	b.callbackView[callback] = view
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

			b.isDebug = false
			b.jsonDebug(update)

			b.handlerUpdate(updateCtx, &update)

			cancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) jsonDebug(update any) {
	if b.isDebug {
		updateByte, err := json.MarshalIndent(update, "", " ")
		if err != nil {
			b.log.Error("%v", err)
		}
		b.log.Info("%s", updateByte)
	}
}

func (b *Bot) handlerUpdate(ctx context.Context, update *tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			b.log.Error("panic recovered: %v, %s", p, string(debug.Stack()))
		}
	}()

	// if write message
	if update.Message != nil {
		b.log.Info("[%s] %s", update.Message.From.UserName, update.Message.Text)

		isProcessing, err := b.isStoreProcessing(ctx, update)
		if err != nil {
			b.log.Error("failed in isStoreProcessing: %v", err)
			handler.HandleError(b.bot, update, err)
			return
		}

		if isProcessing {
			return
		}

		if err := b.userService.CreateUserIFNotExist(ctx, userUpdateToModel(update)); err != nil {
			b.log.Error("userService.CreateUserIfNotExist: failed to create user: %v", err)
			return
		}

		// создание вопроса
		if update.Message.Text != "/admin" && update.Message.Text != "/start" && update.Message.Text != "/cancel" {
			go b.userService.CreateQuestion(context.Background(), update.FromChat().ID, update.Message.Text)
			return
		}

		var view ViewFunc

		cmd := update.Message.Command()

		cmdView, ok := b.cmdView[cmd]
		if !ok {
			return
		}

		view = cmdView

		if err := view(ctx, b.bot, update); err != nil {
			b.log.Error("failed to handle VIEW update: %v", err)
			handler.HandleError(b.bot, update, err)
			return
		}
		//  if press button
	} else if update.CallbackQuery != nil {
		b.log.Info("[%s] %s", update.CallbackQuery.From.UserName, update.CallbackData())

		var callback ViewFunc

		err, callbackView := b.CallbackStrings(update.CallbackData())
		if err != nil {
			b.log.Error("%v", err)
			return
		}

		callback = callbackView

		if err := callback(ctx, b.bot, update); err != nil {
			b.log.Error("failed to handle CALLBACK update: %v", err)
			handler.HandleError(b.bot, update, err)
			return
		}
	}
}
