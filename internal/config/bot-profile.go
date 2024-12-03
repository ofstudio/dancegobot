package config

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
)

var botProfile models.Profile

// BotProfile - returns bot profile
func BotProfile() models.Profile {
	return botProfile
}

// SetBotProfile - sets bot profile
func SetBotProfile(u *tele.User) {
	if u != nil {
		botProfile = models.NewProfile(*u)
	}
}
