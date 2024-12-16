package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestDeeplink_String(t *testing.T) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	url := Deeplink{
		Action:  models.SessionSignup,
		EventID: "eventID",
		Role:    models.RoleLeader,
	}.String()
	assert.Regexp(t, `^https://t.me/my_bot\?start=[a-zA-Z0-9]{4}-signup-eventID-leader$`, url)
}

func TestDeeplinkParsePayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected *Deeplink
		err      bool
	}{
		{
			name:    "valid signup",
			payload: "AD6s-signup-huw8HMZsOp3-leader",
			expected: &Deeplink{
				Action:  models.SessionSignup,
				EventID: "huw8HMZsOp3",
				Role:    models.RoleLeader,
			},
			err: false,
		},
		{
			name:     "invalid action",
			payload:  "AD6s-invalid-huw8HMZsOp3-leader",
			expected: nil,
			err:      true,
		},
		{
			name:     "missing params",
			payload:  "AD6s-signup-huw8HMZsOp3",
			expected: nil,
			err:      true,
		},
		{
			name:     "too few parts",
			payload:  "AD6s-signup",
			expected: nil,
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeeplinkParsePayload(tt.payload)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
