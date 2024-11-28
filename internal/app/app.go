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
	cfg config.Config
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

func (a *App) Start(ctx context.Context) error {

	// 1. Create a new Telegram bot
	bot, err := telegram.NewBot(a.cfg.Bot, a.log)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	views.SetBotName(bot.Me.Username)
	a.log.Info("Bot created", "username", bot.Me.Username, "", a.cfg.Bot)

	// 2. Connect the database and store
	db, err := store.NewSQLite(a.cfg.DB.Filepath, a.cfg.DB.Version)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	a.log.Info("Database connected", "", a.cfg.DB)
	s := store.NewStore(db)
	defer s.Close()

	// 3. Initialize services
	srv := services.NewServices(a.cfg.Settings, s, views.Render(bot), views.Notify(bot)).
		WithLogger(a.log)

	// 4. Initialize middleware and handlers
	m := telegram.NewMiddleware(bot.Me).WithLogger(a.log)
	h := telegram.NewHandlers(a.cfg.Settings, srv.Event, srv.User).WithLogger(a.log)

	// 5. Set up bot middleware and handlers
	bot.Use(m.Context(ctx))
	bot.Use(m.Trace())
	bot.Use(m.Logger())
	bot.Use(m.ChatLink())

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
