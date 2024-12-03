package deeplink

import (
	"strings"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

const dlSeparator = '-'

// New creates a new deeplink to the bot.
//
// Deeplinks format:
//
//	[4 random characters]-[models.SessionAction]-[Param 1]-[Param 2]-[Param...]
//
// Example: sign up for the event with the ID "huw8HMZsOp3" as a leader
//
//	https://t.me/dancegobot?start=AD6s-signup-huw8HMZsOp3-leader
func New(action models.SessionAction, params ...string) string {
	return "https://t.me/" + config.BotProfile().Username + "?start=" +
		randtoken.New(4) + "-" +
		string(action) + "-" +
		strings.Join(params, "-")
}
