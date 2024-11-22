package models

import tele "gopkg.in/telebot.v4"

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

// FullName returns a full name of the Telegram profile.
func (p Profile) FullName() string {
	if p.LastName != "" {
		return p.FirstName + " " + p.LastName
	}
	return p.FirstName
}
