package config

import (
	"time"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
)

// Default returns default configuration
func Default() Config {
	return Config{

		// Bot default configuration
		Bot: Bot{
			ApiURL:         tele.DefaultApiURL,
			UseWebhook:     false,
			WebhookListen:  ":8080",
			RPS:            30,
			Timeout:        30 * time.Second,
			AllowedUpdates: []string{"message", "channel_post", "inline_query", "chosen_inline_result", "callback_query"},
			CommandsPrivate: []tele.Command{
				{Text: "start", Description: locale.CmdDescriptionStart},
				{Text: "settings", Description: locale.CmdDescriptionSettings},
			},
		},

		// Database default configuration
		DB: DB{
			Version: 2,
		},

		// Application default settings
		Settings: Settings{
			EventIDLen:            12,
			EventTextMaxLen:       2048,
			DancerNameMaxLen:      64,
			RendererRepeats:       []time.Duration{2 * time.Second, 10 * time.Second, time.Minute, 1 * time.Hour},
			ReRenderOnStartup:     12 * time.Hour,
			DraftCleanupOlderThan: 24 * time.Hour,
			DraftCleanupEvery:     24 * time.Hour,
		},
	}
}
