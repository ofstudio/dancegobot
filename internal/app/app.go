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
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/noplog"
)

type App struct {
	cfg             config.Config
	Bot             *tele.Bot
	Store           store.Store
	EventService    *services.EventService
	UserService     *services.UserService
	NotifierService *services.NotifierService
	RenderService   *services.RenderService
	log             *slog.Logger
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

// Init initializes the application without starting the Telegram poller.
func (a *App) Init(ctx context.Context) error {
	var err error

	// Create a new Telegram bot
	a.Bot, err = telegram.NewBot(a.cfg.Bot, a.log)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	config.SetBotProfile(a.Bot.Me)
	a.log.Info("Bot created", "", config.BotProfile(), "", a.cfg.Bot)

	// Connect the database and store
	db, err := store.NewSQLite(a.cfg.DB.Filepath, a.cfg.DB.Version)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	a.log.Info("Database connected", "", a.cfg.DB)
	a.Store = store.NewSQLiteStore(db)

	// Initialize services
	a.RenderService = services.
		NewRenderService(a.cfg.Settings, a.Store, views.Render(a.Bot)).
		WithLogger(a.log)
	a.NotifierService = services.
		NewNotifierService(a.cfg.Settings, a.Store, views.Notify(a.Bot)).
		WithLogger(a.log)
	a.EventService = services.
		NewEventService(a.cfg.Settings, a.Store, a.RenderService, a.NotifierService).
		WithLogger(a.log)
	a.UserService = services.
		NewUserService(a.cfg.Settings, a.Store).
		WithLogger(a.log)

	// Start background tasks
	a.EventService.Start(ctx)
	a.RenderService.Start(ctx)

	// Wire middleware and handlers
	a.wire(ctx)

	return nil
}

// Start starts the application.
// Application stops when the context is done.
func (a *App) Start(ctx context.Context) error {
	if err := a.Init(ctx); err != nil {
		return err
	}
	defer a.Close()

	// Start the bot
	go a.Bot.Start()
	a.log.Info("Bot started")

	// Wait for the context to be done
	<-ctx.Done()

	// Stop the bot
	a.Bot.Stop()
	a.log.Info("Bot stopped")

	return nil
}

// Close releases application resources.
func (a *App) Close() {
	if a.Store != nil {
		a.Store.Close()
		a.Store = nil
	}
}
