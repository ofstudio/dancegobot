package views

import (
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventPostAnswerEscapesMessageText(t *testing.T) {
	ctx := &answerContext{}
	event := &models.Event{
		ID:      "event_1",
		Caption: "Танцы <tag> & friends",
	}

	require.NoError(t, EventPostAnswer(ctx, event, "thumb"))
	require.NotNil(t, ctx.resp)
	require.Len(t, ctx.resp.Results, 1)

	result, ok := ctx.resp.Results[0].(*tele.ArticleResult)
	require.True(t, ok)
	require.Equal(t, event.Caption, result.Title)

	content, ok := result.Content.(*tele.InputTextMessageContent)
	require.True(t, ok)
	require.Equal(t, "Танцы &lt;tag&gt; &amp; friends", content.Text)
	require.Equal(t, tele.ModeHTML, content.ParseMode)
}

type answerContext struct {
	tele.Context
	resp *tele.QueryResponse
}

func (c *answerContext) Answer(resp *tele.QueryResponse) error {
	c.resp = resp
	return nil
}
