package deeplink

import (
	"strings"

	"github.com/ofstudio/dancegobot/helpers"
	"github.com/ofstudio/dancegobot/internal/models"
)

const (
	dlSeparator = '-'
	dlMaxLength = 13 + 32 + 7 + 64 // https://t.me/ + username + ?start= + payload
)

var (
	dlBegin = []byte("https://t.me/")
	dlQuery = []byte("?start=")
)

// New creates a new deeplink to the bot.
//
// Deeplinks format:
//
//	[4 random characters]-[models.SessionAction]-[Param 1]-[Param 2]-[Param...]
//
// Example: sign up for the event with the ID "huw8HMZsOp3" as a leader
//
//	https://t.me/dancegobot?start=AD6s-signup-huw8HMZsOp3-leader
func New(botName string, action models.SessionAction, params ...string) string {
	return "https://t.me/" + botName + "?start=" +
		helpers.RandToken(4) + "-" +
		string(action) + "-" +
		strings.Join(params, "-")
}
