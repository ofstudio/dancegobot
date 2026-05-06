package handlers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventQueryParseLimit(t *testing.T) {
	h := &Handlers{}
	defaultSettings := models.EventSettings{Limit: 12}

	tests := []struct {
		name      string
		text      string
		wantText  string
		wantLimit int
	}{
		{
			name:      "valid limit at the end",
			text:      "Limited event /5",
			wantText:  "Limited event",
			wantLimit: 5,
		},
		{
			name:      "minimum limit",
			text:      "Limited event /1",
			wantText:  "Limited event",
			wantLimit: 1,
		},
		{
			name:      "maximum limit",
			text:      "Limited event /99",
			wantText:  "Limited event",
			wantLimit: 99,
		},
		{
			name:      "valid limit in the middle",
			text:      "Limited /5 event",
			wantText:  "Limited event",
			wantLimit: 5,
		},
		{
			name:      "date is not a limit",
			text:      "Class 3/4/2026",
			wantText:  "Class 3/4/2026",
			wantLimit: 12,
		},
		{
			name:      "zero is ordinary text",
			text:      "Limited event /0",
			wantText:  "Limited event /0",
			wantLimit: 12,
		},
		{
			name:      "hundred is ordinary text",
			text:      "Limited event /100",
			wantText:  "Limited event /100",
			wantLimit: 12,
		},
		{
			name:      "slash without leading space is ordinary text",
			text:      "Limited/5 event",
			wantText:  "Limited/5 event",
			wantLimit: 12,
		},
		{
			name:      "shortcut without leading text is ordinary text",
			text:      "/5 Limited event",
			wantText:  "/5 Limited event",
			wantLimit: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotSettings := h.eventQueryParse(tt.text, defaultSettings)
			require.Equal(t, tt.wantText, gotText)
			require.Equal(t, tt.wantLimit, gotSettings.Limit)
		})
	}
}
