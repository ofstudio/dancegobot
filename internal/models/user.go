package models

import (
	"log/slog"
	"time"
)

// User - is a user of the bot
type User struct {
	Profile   Profile
	Session   Session
	Settings  UserSettings
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserSettings - is a user settings
type UserSettings struct {
	Event EventSettings `json:"event"` // Default settings for new events created by user
}

// LogValue returns a string representation of the UserSettings
func (s UserSettings) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("event", s.Event),
	)
}
