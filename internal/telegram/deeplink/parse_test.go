package deeplink

import (
	"github.com/ofstudio/dancegobot/internal/models"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePayload(t *testing.T) {
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
			action, params, err := ParsePayload(tt.payload)
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
