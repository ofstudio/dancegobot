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
	Events EventsSettings `json:"events"` // Settings for events created by user
}

// EventsSettings - is a settings for events created by user
type EventsSettings struct {
	DisableChooseSingle bool `json:"disable_choose_single,omitempty"` // Disable choose specific single dancer from the list
}
