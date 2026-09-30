package views

import (
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventPostAnswerPersonalCache(t *testing.T) {
	ctx := &answerContext{}
	event := &models.Event{
		ID:      "event_1",
		Caption: "Dance class",
	}

	require.NoError(t, EventPostAnswer(ctx, event, "thumb"))
	require.NotNil(t, ctx.resp)
	require.True(t, ctx.resp.IsPersonal)
	require.Equal(t, 1, ctx.resp.CacheTime)
	require.Len(t, ctx.resp.Results, 1)

	result, ok := ctx.resp.Results[0].(*tele.ArticleResult)
	require.True(t, ok)
	require.Equal(t, event.ID, result.ID)
	require.NotNil(t, result.ReplyMarkup)
	require.Len(t, result.ReplyMarkup.InlineKeyboard, 1)
	require.Len(t, result.ReplyMarkup.InlineKeyboard[0], 2)
	require.Equal(t, BtnEventSignupCb.Unique, result.ReplyMarkup.InlineKeyboard[0][0].Unique)
	require.Equal(t, BtnEventSignupCb.Unique, result.ReplyMarkup.InlineKeyboard[0][1].Unique)
	require.Contains(t, result.ReplyMarkup.InlineKeyboard[0][0].Data, event.ID+"|leader|")
	require.Contains(t, result.ReplyMarkup.InlineKeyboard[0][1].Data, event.ID+"|follower|")
}

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
