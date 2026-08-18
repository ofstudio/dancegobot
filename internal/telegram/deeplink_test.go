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
	deeplink := Deeplink{
		Action:  models.SessionSignup,
		EventID: "eventID",
		Role:    models.RoleLeader,
	}
	want := "https://t.me/my_bot?start=v1-signup-eventID-leader"

	assert.Equal(t, want, deeplink.String())
	assert.Equal(t, want, deeplink.String())
}

func TestDeeplink_Subscribe(t *testing.T) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	deeplink := Deeplink{Action: models.SessionSubscribe, EventID: "eventID"}
	want := "https://t.me/my_bot?start=v1-subscribe-eventID"

	assert.Equal(t, want, deeplink.String())
	parsed, err := DeeplinkParse(want)
	assert.NoError(t, err)
	assert.Equal(t, &deeplink, parsed)
}

func TestDeeplinkParsePayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected *Deeplink
		err      bool
	}{
		{
			name:    "current version signup",
			payload: "v1-signup-huw8HMZsOp3-leader",
			expected: &Deeplink{
				Action:  models.SessionSignup,
				EventID: "huw8HMZsOp3",
				Role:    models.RoleLeader,
			},
			err: false,
		},
		{
			name:    "legacy random nonce signup",
			payload: "AD6s-signup-huw8HMZsOp3-leader",
			expected: &Deeplink{
				Action:  models.SessionSignup,
				EventID: "huw8HMZsOp3",
				Role:    models.RoleLeader,
			},
			err: false,
		},
		{
			name:    "subscribe",
			payload: "v1-subscribe-huw8HMZsOp3",
			expected: &Deeplink{
				Action:  models.SessionSubscribe,
				EventID: "huw8HMZsOp3",
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
