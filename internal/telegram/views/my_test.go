package views

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
)

func TestBtnMySceneSubscriptionBelowChatLink(t *testing.T) {
	reg := models.Registration{Event: &models.Event{
		ID:       "event_id",
		Settings: models.EventSettings{Closed: true},
		Post: &models.Post{
			Chat:          &models.Chat{ID: -1001234567890, Type: models.ChatSuper},
			ChatMessageID: 42,
		},
	}}

	rm := btnMyScene(reg, false, services.SubscriptionStatus{Available: true}, 0, 0)
	require.Len(t, rm.InlineKeyboard, 2)
	require.Equal(t, locale.BtnChatLink, rm.InlineKeyboard[0][0].Text)
	require.Equal(t, locale.BtnSubscribe, rm.InlineKeyboard[1][0].Text)
}
