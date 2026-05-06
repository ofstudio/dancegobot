package views

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventSettingsSceneEscapesCaption(t *testing.T) {
	ctx := &editOrSendContext{}
	event := &models.Event{
		ID:        "event_1",
		Caption:   "Settings <tag> & owner",
		CreatedAt: time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
	}

	require.NoError(t, EventSettingsScene(ctx, event, 0))

	text, ok := ctx.what.(string)
	require.True(t, ok)
	require.Contains(t, text, "Settings &lt;tag&gt; &amp; owner")
	require.NotContains(t, text, "Settings <tag> & owner")
}

type editOrSendContext struct {
	tele.Context
	what interface{}
}

func (c *editOrSendContext) EditOrSend(what interface{}, opts ...interface{}) error {
	c.what = what
	return nil
}
