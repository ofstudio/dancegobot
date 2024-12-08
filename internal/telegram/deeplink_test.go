package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

func Test_deeplink(t *testing.T) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	deeplink := deeplink("signup", "eventID", "leader")
	assert.Regexp(t, `^https://t.me/my_bot\?start=[a-zA-Z0-9]{4}-signup-eventID-leader$`, deeplink)
}

func Benchmark_deeplink(b *testing.B) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	for i := 0; i < b.N; i++ {
		_ = deeplink("event_signup", randtoken.New(12), "leader")
	}
}

func Test_deeplinkParsePayload(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected models.SessionAction
		params   []string
		err      bool
	}{
		{
			name:     "valid signup",
			payload:  "AD6s-signup-huw8HMZsOp3-leader",
			expected: models.SessionSignup,
			params:   []string{"huw8HMZsOp3", "leader"},
			err:      false,
		},
		{
			name:    "invalid action",
			payload: "AD6s-invalid-huw8HMZsOp3-leader",
			err:     true,
		},
		{
			name:    "missing params",
			payload: "AD6s-signup-huw8HMZsOp3",
			err:     true,
		},
		{
			name:    "too few parts",
			payload: "AD6s-signup",
			err:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, params, err := deeplinkParsePayload(tt.payload)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, action)
				assert.Equal(t, tt.params, params)
			}
		})
	}
}
