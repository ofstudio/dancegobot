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

type UserSettings struct {
}
