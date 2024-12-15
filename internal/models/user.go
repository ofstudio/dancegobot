package models

import (
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
	Events UserEventSettings `json:"events"` // Settings for events created by user
}

// UserEventSettings - is a settings for events created by user
type UserEventSettings struct {
	AutoPairing bool `json:"auto_pairing,omitempty"` // Automatically pair single dancers
}
