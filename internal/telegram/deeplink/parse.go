package deeplink

import (
	"fmt"
	"strings"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

// Parse parses and validates deeplink.
// Returns the user action and parameters.
func Parse(deeplink string) (models.SessionAction, []string, error) {
	if !strings.HasPrefix(deeplink, "https://t.me/"+config.BotProfile().Username+"?start=") {
		return "", nil, errInvalid(deeplink)
	}

	parts := strings.Split(deeplink, "?start=")
	if len(parts) < 2 {
		return "", nil, errInvalid(deeplink)
	}

	return ParsePayload(parts[1])
}

// ParsePayload parses and validates deeplink payload.
// Returns the user action and parameters.
func ParsePayload(payload string) (models.SessionAction, []string, error) {

	parts := strings.Split(payload, string(dlSeparator))
	if len(parts) < 2 {
		return "", nil, errInvalid(payload)
	}

	action := models.SessionAction(parts[1])
	params := parts[2:]

	switch action {
	case models.SessionSignup:
		if len(params) != 2 {
			return "", nil, errInvalid(payload)
		}
	default:
		return "", nil, errInvalid(payload)
	}

	return action, params, nil
}

func errInvalid(payload string) error {
	return fmt.Errorf("invalid deeplink payload: %q", payload)
}
