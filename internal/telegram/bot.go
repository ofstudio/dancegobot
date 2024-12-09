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

// NewBot creates a new telegram bot.
func NewBot(cfg config.Bot, log *slog.Logger) (*tele.Bot, error) {
	poller := func(cfg config.Bot) tele.Poller {
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

	if log == nil {
		log = noplog.Logger()
	}

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

	return bot, nil
}

// onError is a bot error handler.
func onError(log *slog.Logger) func(err error, c tele.Context) {
	return func(err error, c tele.Context) {

		// Ignore true result errors.
		// https://github.com/tucnak/telebot/issues/758
		// telebot v4.0.0-beta.4
		if errors.Is(err, tele.ErrTrueResult) {
			return
		}

		if c == nil {
			log.Error(err.Error())
		} else {
			log.Error("[bot] "+err.Error(), telelog.Attr(c))
		}
	}
}

var scopePrivate = tele.CommandScope{Type: tele.CommandScopeAllPrivateChats}
