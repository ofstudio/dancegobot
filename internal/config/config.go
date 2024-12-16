package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	tele "gopkg.in/telebot.v4"
)

// Config is application configuration
type Config struct {
	Bot      // Telegram bot configuration
	DB       // SQLite database configuration
	Settings // Application settings
}

// DB is SQLite database configuration
type DB struct {
	Filepath string `env:"DB_FILEPATH,required"` // Path to database file
	Version  uint   // Required database schema version
}

// Settings - application settings
type Settings struct {
	QueryThumbUrl         string          `env:"THUMBNAIL_URL"` // URL for thumbnail image for query answer
	EventIDLen            int             // Length of event ID
	EventTextMaxLen       int             // Maximum length for event text in runes
	DancerNameMaxLen      int             // Maximum length for dancer name in runes
	RendererRepeats       []time.Duration // Time intervals for event rendering repeats
	ReRenderOnStartup     time.Duration   // Re-render on startup the recent events that were updated not older than this duration
	DraftCleanupOlderThan time.Duration   // Cleanup event drafts that were created older than this duration
	DraftCleanupEvery     time.Duration   // Cleanup event drafts every this duration since startup
}

// Bot is Telegram bot configuration
type Bot struct {
	ApiURL           string         `env:"BOT_API_URL"`
	Token            string         `env:"BOT_TOKEN,required,unset"`
	UseWebhook       bool           `env:"BOT_USE_WEBHOOK"`
	WebhookListen    string         `env:"BOT_WEBHOOK_LISTEN"`
	WebhookPublicURL string         `env:"BOT_WEBHOOK_PUBLIC_URL"`
	RPS              int            // Requests per second
	Timeout          time.Duration  // Poller and http-client timeouts
	AllowedUpdates   []string       // Allowed update types
	CommandsPrivate  []tele.Command // Bot commands for private chats
}

// Load loads configuration from [Default] and environment variables
func Load() (Config, error) {
	c := Default()
	if err := env.Parse(&c); err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return c, nil
}
