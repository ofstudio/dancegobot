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

// String returns the deep link URL.
func (d Deeplink) String() string {
	url := "https://t.me/" + config.BotProfile().Username + "?start=" + randtoken.New(4) + dlSeparator

	switch d.Action {
	case models.SessionSignup:
		url += string(d.Action) + dlSeparator + d.EventID + dlSeparator + string(d.Role)
	default:
	}
	return url
}

//	func deeplink(action models.SessionAction, params ...string) string {
//		return "https://t.me/" + config.BotProfile().Username + "?start=" +
//			randtoken.New(4) + "-" +
//			string(action) + "-" +
//			strings.Join(params, "-")
//	}
//
// // deeplinkParse parses and validates deep link.
// // Returns the user action and parameters.
//
//	func deeplinkParse(deeplink string) (models.SessionAction, []string, error) {
//		if !strings.HasPrefix(deeplink, "https://t.me/"+config.BotProfile().Username+"?start=") {
//			return "", nil, errDeepLinkInvalid(deeplink)
//		}
//
//		parts := strings.Split(deeplink, "?start=")
//		if len(parts) < 2 {
//			return "", nil, errDeepLinkInvalid(deeplink)
//		}
//
//		return deeplinkParsePayload(parts[1])
//	}
//
// // deeplinkParsePayload parses and validates deeplink payload.
// // Returns the user action and parameters.
// func deeplinkParsePayload(payload string) (models.SessionAction, []string, error) {
//
//		parts := strings.Split(payload, dlSeparator)
//		if len(parts) < 2 {
//			return "", nil, errDeepLinkInvalid(payload)
//		}
//
//		action := models.SessionAction(parts[1])
//		params := parts[2:]
//
//		switch action {
//		case models.SessionSignup:
//			if len(params) != 2 {
//				return "", nil, errDeepLinkInvalid(payload)
//			}
//		default:
//			return "", nil, errDeepLinkInvalid(payload)
//		}
//
//		return action, params, nil
//	}
func errDeeplink(url string) error {
	return fmt.Errorf("invalid deeplink: %q", url)
}
func errPayload(payload string) error {
	return fmt.Errorf("invalid deeplink payload: %q", payload)
}
