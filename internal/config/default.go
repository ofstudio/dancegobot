package config

import (
	"time"

	tele "gopkg.in/telebot.v4"
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
		},

		// Database default configuration
		DB: DB{
			Version: 2,
		},

		// Application default settings
		Settings: Settings{
			EventIDLen:       12,
			EventTextMaxLen:  2048,
			DancerNameMaxLen: 64,
			RendererRepeats:  []time.Duration{2 * time.Second, 10 * time.Second, time.Minute, 1 * time.Hour},
			ChatMessageDelay: 500 * time.Millisecond,
		},
	}
}
