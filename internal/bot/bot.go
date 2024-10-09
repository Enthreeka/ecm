package bot

import (
	"context"
	"github.com/Enthreeka/tg-question-bot/internal/config"
	"github.com/Enthreeka/tg-question-bot/internal/handler/callback"
	"github.com/Enthreeka/tg-question-bot/internal/handler/middleware"
	"github.com/Enthreeka/tg-question-bot/internal/handler/tgbot"
	"github.com/Enthreeka/tg-question-bot/internal/handler/view"
	"github.com/Enthreeka/tg-question-bot/internal/repo"
	service "github.com/Enthreeka/tg-question-bot/internal/usecase"
	"github.com/Enthreeka/tg-question-bot/pkg/excel"
	store "github.com/Enthreeka/tg-question-bot/pkg/local_storage"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	"github.com/Enthreeka/tg-question-bot/pkg/postgres"
	customMsg "github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

const (
	PostgresMaxAttempts = 5
)

type Bot struct {
	bot           *tgbotapi.BotAPI
	psql          *postgres.Postgres
	store         *store.Store
	excel         *excel.Excel
	cfg           *config.Config
	log           *logger.Logger
	tgMsg         *customMsg.TelegramMsg
	callbackStore *store.CallbackStorage

	userService service.UserService
	userRepo    repo.UserRepo

	callbackUser callback.CallbackUser
	viewGeneral  *view.ViewGeneral
}

func NewBot() *Bot {
	return &Bot{}
}

func (b *Bot) initExcel() {
	b.excel = excel.NewExcel(b.log)
}

func (b *Bot) initHandler() {
	b.viewGeneral = view.NewViewGeneral(b.log, b.tgMsg, b.psql)

	callbackUser, err := callback.NewCallbackUser(b.userService, b.log, b.store, b.tgMsg, b.psql, b.excel)
	if err != nil {
		log.Fatal(err)
	}
	b.callbackUser = callbackUser

	b.log.Info("Initializing handler")
}

func (b *Bot) initUsecase() {
	userService, err := service.NewUserService(b.userRepo, b.log, b.tgMsg, b.psql)
	if err != nil {
		b.log.Fatal("Failed to initialize user service")
	}
	b.userService = userService

	b.log.Info("Initializing usecase")
}

func (b *Bot) initRepo() {
	userRepo, err := repo.NewUserRepo(b.psql)
	if err != nil {
		log.Fatal("Failed to initialize user repo")
	}

	b.userRepo = userRepo

	b.log.Info("Initializing repo")
}

func (b *Bot) initMessage() {
	b.tgMsg = customMsg.NewMessageSetting(b.bot, b.log)

	b.log.Info("Initializing message")
}

func (b *Bot) initPostgres(ctx context.Context) {
	psql, err := postgres.New(ctx, PostgresMaxAttempts, b.cfg.Postgres.URL)
	if err != nil {
		b.log.Fatal("failed to connect PostgreSQL: %v", err)
	}
	b.psql = psql

	b.log.Info("Initializing postgres")
}

func (b *Bot) initConfig() {
	cfg, err := config.New()
	if err != nil {
		b.log.Fatal("failed load config: %v", err)
	}
	b.cfg = cfg

	b.log.Info("Initializing config")
}

func (b *Bot) initLogger() {
	b.log = logger.New()

	b.log.Info("Initializing logger")
}

func (b *Bot) initStore() {
	b.store = store.NewStore()

	b.log.Info("Initializing store")
}

func (b *Bot) initCallbackStorage() {
	b.callbackStore = store.NewCallbackStorage()

	b.log.Info("Initializing callback storage")
}

func (b *Bot) initTelegramBot() {
	bot, err := tgbotapi.NewBotAPI(b.cfg.Telegram.Token)
	if err != nil {
		b.log.Fatal("failed to load token %v", err)
	}
	bot.Debug = false
	b.bot = bot

	b.log.Info("Initializing telegram bot")
	b.log.Info("Authorized on account %s", bot.Self.UserName)
}

func (b *Bot) initialize(ctx context.Context) {
	b.initLogger()
	b.initExcel()
	b.initConfig()
	b.initTelegramBot()
	b.initStore()
	b.initCallbackStorage()
	b.initPostgres(ctx)
	b.initMessage()
	b.initRepo()
	b.initUsecase()
	b.initHandler()

}

func (b *Bot) Run(ctx context.Context) {
	startBot := time.Now()
	b.initialize(ctx)
	newBot, err := tgbot.NewBot(b.bot, b.log, b.store, b.tgMsg, b.userService, b.callbackStore)
	if err != nil {
		b.log.Fatal("failed go create new bot: ", err)
	}
	defer b.psql.Close()

	newBot.RegisterCommandView("start", b.viewGeneral.CallbackStartUser())

	newBot.RegisterCommandView("admin", middleware.AdminMiddleware(b.userService, b.viewGeneral.CallbackStartAdminPanel()))

	newBot.RegisterCommandCallback("bot_setting", middleware.AdminMiddleware(b.userService, b.callbackUser.QuestionSettings()))
	newBot.RegisterCommandCallback("main_menu", middleware.AdminMiddleware(b.userService, b.callbackUser.MainMenu()))
	newBot.RegisterCommandCallback("user_setting", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminRoleSetting()))
	newBot.RegisterCommandCallback("admin_look_up", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminLookUp()))
	newBot.RegisterCommandCallback("admin_delete_role", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminDeleteRole()))
	newBot.RegisterCommandCallback("admin_set_role", middleware.AdminMiddleware(b.userService, b.callbackUser.AdminSetRole()))

	b.log.Info("Initialize bot took [%f] seconds", time.Since(startBot).Seconds())
	if err := newBot.Run(ctx); err != nil {
		b.log.Fatal("failed to run Telegram Bot: %v", err)
	}
}
