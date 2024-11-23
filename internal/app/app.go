package app

import (
	"context"
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/internal/telegram"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

type App struct {
	cfg config.Config
	log *slog.Logger
}

func New(cfg config.Config) *App {
	return &App{
		cfg: cfg,
		log: noplog.Logger(),
	}
}

func (a *App) WithLogger(log *slog.Logger) *App {
	a.log = log
	return a
}

func (a *App) Start(ctx context.Context) error {

	// 1. Create a new Telegram bot
	bot, err := telegram.NewBot(a.cfg.Bot, a.log)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	a.cfg.Settings.BotName = telegram.BotName(bot)
	a.log.Info("Bot created", "username", a.cfg.Settings.BotName, "", a.cfg.Bot)

	// 2. Initialize the database and store
	db, err := store.NewSQLite("./playground/dev.db", a.cfg.DB.Version)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.log.Info("Database initialized", "", a.cfg.DB)
	s := store.NewStore(db)
	defer s.Close()

	// 3. Initialize services
	srv := services.NewServices(a.cfg.Settings, s, bot).WithLogger(a.log)

	// 4. Initialize middleware and handlers
	m := telegram.NewMiddleware().WithLogger(a.log)
	h := telegram.NewHandlers(a.cfg.Settings, srv.Event, srv.User).WithLogger(a.log)

	// 5. Set up bot middleware and handlers
	bot.Use(m.Context(ctx))
	bot.Use(m.Trace())
	bot.Use(m.Logger())
	bot.Use(m.PassPrivateMessages())

	bot.Handle("/start", h.Start)
	bot.Handle("/partner", h.Partner)
	bot.Handle(tele.OnText, h.Text)
	bot.Handle(tele.OnUserShared, h.UserShared)
	bot.Handle(tele.OnQuery, h.Query)
	bot.Handle(tele.OnInlineResult, h.InlineResult)

	// 6. Start the bot
	go bot.Start()
	a.log.Info("Bot started")

	// 7. Wait for the context to be done
	<-ctx.Done()

	// 8. Stop the bot
	bot.Stop()
	a.log.Info("Bot stopped")

	return nil
}
