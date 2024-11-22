package config

import (
	"fmt"
	"time"

	env "github.com/caarlos0/env/v11"
)

// Config is application configuration
type Config struct {
	Bot      // Telegram bot configuration
	DB       // SQLite database configuration
	Settings // Application settings
}

// DB is SQLite database configuration
type DB struct {
	Filepath        string `env:"DB_FILEPATH,required"` // Path to database file
	RequiredVersion uint   // Required database schema version
}

// Settings - application settings
type Settings struct {
	EventIDLen       int             // Length of event ID
	EventTextMaxLen  int             // Maximum length for event text in runes
	DancerNameMaxLen int             // Maximum length for dancer name in runes
	QueryThumbUrl    string          `env:"THUMBNAIL_URL"` // URL for thumbnail image for query answer
	RendererRepeats  []time.Duration // Time intervals for event rendering repeats
}

// Bot is Telegram bot configuration
type Bot struct {
	ApiURL           string        `env:"BOT_API_URL"`
	Token            string        `env:"BOT_TOKEN,required,unset"`
	UseWebhook       bool          `env:"BOT_USE_WEBHOOK"`
	WebhookListen    string        `env:"BOT_WEBHOOK_LISTEN"`
	WebhookPublicURL string        `env:"BOT_WEBHOOK_PUBLIC_URL"`
	RPS              int           // Requests per second
	Timeout          time.Duration // Poller and http-client timeouts
	AllowedUpdates   []string      // Allowed update types
}

// NewConfig loads configuration from [Default] and environment variables
func NewConfig() (Config, error) {
	c := Default
	if err := env.Parse(&c); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrEnvParse, err)
	}
	return c, nil
}
