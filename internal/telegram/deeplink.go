package telegram

import (
	"fmt"
	"strings"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

const dlSeparator = "-"

// Deeplink - deep link to the bot.
//
// Deeplinks format:
//
//	[4 random characters]-[models.SessionAction]-[Param 1]-[Param 2]-[Param...]
//
// Example: sign up for the event with the ID "huw8HMZsOp3" as a leader
//
//	https://t.me/dancegobot?start=AD6s-signup-huw8HMZsOp3-leader
//
// More info: https://core.telegram.org/api/links#bot-links
type Deeplink struct {
	Action  models.SessionAction
	EventID string
	Role    models.Role
}

// DeeplinkParse parses the deeplink from the URL.
func DeeplinkParse(url string) (*Deeplink, error) {
	if !strings.HasPrefix(url, "https://t.me/"+config.BotProfile().Username+"?start=") {
		return nil, errDeeplink(url)
	}

	parts := strings.Split(url, "?start=")
	if len(parts) < 2 {
		return nil, errDeeplink(url)
	}

	return DeeplinkParsePayload(parts[1])
}

// DeeplinkParsePayload parses the deeplink from 'start' parameter in the URL.
func DeeplinkParsePayload(payload string) (*Deeplink, error) {
	parts := strings.Split(payload, dlSeparator)
	if len(parts) < 2 {
		return nil, errPayload(payload)
	}

	action := models.SessionAction(parts[1])
	params := parts[2:]

	switch action {
	case models.SessionSignup:
		if len(params) < 2 {
			return nil, errPayload(payload)
		}
		return &Deeplink{
			Action:  action,
			EventID: params[0],
			Role:    models.Role(params[1]),
		}, nil
	default:
		return nil, errPayload(payload)
	}
}

// String returns the deep link URL string.
func (d Deeplink) String() string {
	url := "https://t.me/" + config.BotProfile().Username + "?start=" + randtoken.New(4) + dlSeparator

	switch d.Action {
	case models.SessionSignup:
		url += string(d.Action) + dlSeparator + d.EventID + dlSeparator + string(d.Role)
	default:
	}
	return url
}

func errDeeplink(url string) error {
	return fmt.Errorf("invalid deeplink: %q", url)
}

func errPayload(payload string) error {
	return fmt.Errorf("invalid deeplink payload: %q", payload)
}
