package app

import (
	"context"
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/repo"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/telegram"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

type App struct {
	cfg config.Config
	srv *services.Services
	log *slog.Logger
}

// New creates a new application with the given configuration.
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

// Start starts the application.
// Application stops when the context is done.
func (a *App) Start(ctx context.Context) error {

	// 1. Create a new Telegram bot
	bot, err := telegram.NewBot(a.cfg.Bot, a.log)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	config.SetBotProfile(bot.Me)
	a.log.Info("Bot created", "", config.BotProfile(), "", a.cfg.Bot)

	// 2. Connect the database and store
	db, err := repo.NewSQLite(a.cfg.DB.Filepath, a.cfg.DB.Version)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	a.log.Info("Database connected", "", a.cfg.DB)
	store := repo.NewSQLiteStore(db)
	defer store.Close()

	// 3. Initialize services
	a.srv = services.NewServices(
		a.cfg.Settings,
		store,
		telegram.RenderPost(bot),
		telegram.Notify(bot),
	).WithLogger(a.log)

	// 4. Start background tasks
	a.srv.Event.Start(ctx)
	a.srv.Render.Start(ctx)

	// 5. Initialize middleware and handlers
	m := telegram.NewMiddleware(a.cfg.Settings, a.srv.Event, a.srv.User).WithLogger(a.log)
	h := telegram.NewHandlers(a.cfg.Settings, a.srv.Event, a.srv.User).WithLogger(a.log)

	// 6. Set up bot middleware and handlers
	bot.Use(m.Context(ctx))
	bot.Use(m.Trace())
	bot.Use(m.Logger())
	bot.Use(m.ChatMessage())
	bot.Use(m.PassPrivateMessages())
	bot.Use(m.User())

	bot.Handle("/start", h.Start)
	bot.Handle("/partner", h.Partner)
	bot.Handle("/settings", h.Settings)

	bot.Handle(tele.OnText, h.Text)
	bot.Handle(tele.OnUserShared, h.UserShared)
	bot.Handle(tele.OnQuery, h.Query)
	bot.Handle(tele.OnInlineResult, h.InlineResult)

	bot.Handle(&telegram.BtnCbSignup, h.CbSignup)
	bot.Handle(&telegram.BtnCbSettingsAutoPair, h.CbSettingsAutoPair)

	// This is needed to handle channel posts
	bot.Handle(tele.OnChannelPost, func(_ tele.Context) error { return nil })

	// 7. Start the bot
	go bot.Start()
	a.log.Info("Bot started")

	// 8. Wait for the context to be done
	<-ctx.Done()

	// 9. Stop the bot
	bot.Stop()
	a.log.Info("Bot stopped")

	return nil
}
