package views

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestSubscriptionPrompt(t *testing.T) {
	tests := []struct {
		name string
		chat *models.Chat
		want string
	}{
		{
			name: "title",
			chat: &models.Chat{Title: "Dance <Practice>"},
			want: "Подписаться на Dance &lt;Practice&gt;?\n\n" +
				"Вы будете получать уведомления о новых мероприятиях в этом чате.",
		},
		{
			name: "username",
			chat: &models.Chat{Username: "dance_practice"},
			want: "Подписаться на @dance_practice?\n\n" +
				"Вы будете получать уведомления о новых мероприятиях в этом чате.",
		},
		{
			name: "unnamed chat",
			chat: &models.Chat{},
			want: locale.SubscriptionPromptChat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, subscriptionPromptText(tt.chat))
		})
	}

	rm := btnSubscriptionPrompt("event_id")
	require.Len(t, rm.InlineKeyboard, 2)
	require.Equal(t, locale.BtnSubscribe, rm.InlineKeyboard[0][0].Text)
	require.Contains(t, rm.InlineKeyboard[0][0].Data, "event_id")
	require.Equal(t, locale.BtnClose, rm.InlineKeyboard[1][0].Text)
}

func TestSubscriptionResultText(t *testing.T) {
	require.Equal(t,
		"Вы подписались на новые мероприятия в Dance.",
		SubscriptionSubscribedText(&models.Chat{Title: "Dance"}),
	)
	require.Equal(t,
		"Вы подписались на новые мероприятия в @dance.",
		SubscriptionSubscribedText(&models.Chat{Username: "dance"}),
	)
	require.Equal(t,
		locale.SubscriptionSubscribedChat,
		SubscriptionSubscribedText(&models.Chat{}),
	)
	require.Equal(t,
		"Вы отписались от Dance.",
		SubscriptionUnsubscribedText(&models.Chat{Title: "Dance"}),
	)
	require.Equal(t,
		locale.SubscriptionUnsubscribedChat,
		SubscriptionUnsubscribedText(nil),
	)
}

func TestSubscriptionUnsubscribedMessage(t *testing.T) {
	chat := &models.Chat{Title: "Dance & Practice"}
	require.Equal(t,
		"Вы отписались от Dance &amp; Practice.",
		subscriptionUnsubscribedMessage(chat, ""),
	)
	require.Equal(t,
		`Вы отписались от Dance &amp; Practice.`+"\n\n"+
			`<a href="https://t.me/test_bot?start=v1-subscribe-event">`+locale.LinkResubscribe+`</a>`,
		subscriptionUnsubscribedMessage(chat, "https://t.me/test_bot?start=v1-subscribe-event"),
	)
}
