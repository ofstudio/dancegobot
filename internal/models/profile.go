package models

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// Profile is a Telegram user profile.
type Profile struct {
	ID        int64  `json:"id"`                  // Telegram user id
	FirstName string `json:"first_name"`          // First name
	LastName  string `json:"last_name,omitempty"` // Last name
	Username  string `json:"username,omitempty"`  // Telegram username
}

func NewProfile(u tele.User) Profile {
	return Profile{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
	}
}

// LogValue implements slog.Valuer interface for Profile model.
func (p Profile) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Int64("id", p.ID),
		slog.String("first_name", p.FirstName),
	}
	if p.LastName != "" {
		attrs = append(attrs, slog.String("last_name", p.LastName))
	}
	if p.Username != "" {
		attrs = append(attrs, slog.String("username", p.Username))
	}
	return slog.GroupValue(attrs...)
}

// FullName returns a full name of the Telegram profile.
func (p Profile) FullName() string {
	if p.LastName != "" {
		return p.FirstName + " " + p.LastName
	}
	return p.FirstName
}
