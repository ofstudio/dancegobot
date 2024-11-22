package deeplink

import (
	"fmt"
	"strings"

	"github.com/ofstudio/dancegobot/internal/models"
)

// Parse parses and validates deeplink payload.
// Returns the user action and parameters.
func Parse(payload string) (models.SessionAction, []string, error) {

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
