package telegram

import (
	"errors"
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/ratelimit"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

var scopePrivate = tele.CommandScope{Type: tele.CommandScopeAllPrivateChats}

// NewBot creates a new telegram bot.
func NewBot(cfg config.Bot, log *slog.Logger) (*tele.Bot, error) {
	bot, err := tele.NewBot(tele.Settings{
		URL:     cfg.ApiURL,
		Token:   cfg.Token,
		Poller:  poller(cfg),
		OnError: onError(log),
		Client:  ratelimit.Client(cfg.RPS, cfg.Timeout),
	})
	if err != nil {
		return nil, err
	}

	// Set bot commands for private chats.
	if err = bot.SetCommands(cfg.CommandsPrivate, scopePrivate); err != nil {
		return nil, fmt.Errorf("failed to set bot commands: %w", err)
	}

	// Remove webhook if in long polling mode.
	if !cfg.UseWebhook {
		if err = bot.RemoveWebhook(); err != nil {
			return nil, fmt.Errorf("failed to remove webhook: %w", err)
		}
	}

	return bot, nil
}

// onError is a bot error handler.
func onError(log *slog.Logger) func(err error, c tele.Context) {
	if log == nil {
		log = noplog.Logger()
	}
	return func(err error, c tele.Context) {

		// Ignore true result errors.
		// https://github.com/tucnak/telebot/issues/758
		// telebot v4.0.0-beta.4
		if errors.Is(err, tele.ErrTrueResult) {
			return
		}

		msg := "[bot] " + err.Error()
		if c == nil {
			log.Error(msg)
		} else {
			log.Error(msg, telelog.Attr(c))
		}
	}
}

// poller returns a poller type based on the configuration.
func poller(cfg config.Bot) tele.Poller {
	if cfg.UseWebhook {
		return &tele.Webhook{
			Listen:         cfg.WebhookListen,
			AllowedUpdates: cfg.AllowedUpdates,
			SecretToken:    randtoken.New(64),
			Endpoint:       &tele.WebhookEndpoint{PublicURL: cfg.WebhookPublicURL},
		}
	}
	return &tele.LongPoller{
		Timeout:        cfg.Timeout,
		AllowedUpdates: cfg.AllowedUpdates,
	}
}
