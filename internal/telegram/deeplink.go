package telegram

import (
	"fmt"
	"strings"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

const dlSeparator = '-'

// deeplink creates a new deep link to the bot.
//
// Deeplinks format:
//
//	[4 random characters]-[models.SessionAction]-[Param 1]-[Param 2]-[Param...]
//
// Example: sign up for the event with the ID "huw8HMZsOp3" as a leader
//
//	https://t.me/dancegobot?start=AD6s-signup-huw8HMZsOp3-leader
func deeplink(action models.SessionAction, params ...string) string {
	return "https://t.me/" + config.BotProfile().Username + "?start=" +
		randtoken.New(4) + "-" +
		string(action) + "-" +
		strings.Join(params, "-")
}

// deeplinkParse parses and validates deep link.
// Returns the user action and parameters.
func deeplinkParse(deeplink string) (models.SessionAction, []string, error) {
	if !strings.HasPrefix(deeplink, "https://t.me/"+config.BotProfile().Username+"?start=") {
		return "", nil, errDeepLinkInvalid(deeplink)
	}

	parts := strings.Split(deeplink, "?start=")
	if len(parts) < 2 {
		return "", nil, errDeepLinkInvalid(deeplink)
	}

	return deeplinkParsePayload(parts[1])
}

// deeplinkParsePayload parses and validates deeplink payload.
// Returns the user action and parameters.
func deeplinkParsePayload(payload string) (models.SessionAction, []string, error) {

	parts := strings.Split(payload, string(dlSeparator))
	if len(parts) < 2 {
		return "", nil, errDeepLinkInvalid(payload)
	}

	action := models.SessionAction(parts[1])
	params := parts[2:]

	switch action {
	case models.SessionSignup:
		if len(params) != 2 {
			return "", nil, errDeepLinkInvalid(payload)
		}
	default:
		return "", nil, errDeepLinkInvalid(payload)
	}

	return action, params, nil
}

func errDeepLinkInvalid(payload string) error {
	return fmt.Errorf("invalid deeplink payload: %q", payload)
}
