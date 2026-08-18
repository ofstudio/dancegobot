package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

func Test_notifyText(t *testing.T) {
	t.Run("TmplNewEvent", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplNewEvent,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t, "🔔 Новая запись в Test Chat.", text.String())
	})

	t.Run("TmplRegisteredWithSingle", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplRegisteredWithSingle,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> зарегистрировался с тобой в паре! 🎉",
			text.String())
	})
	t.Run("TmplRegisteredWithSingle_waitlist", func(t *testing.T) {
		payload := testPayload
		payload.WaitList = true
		n := &models.Notification{
			TmplCode: models.TmplRegisteredWithSingle,
			Payload:  payload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> зарегистрировался с тобой в паре! 🎉"+
				"\n\nВаша пара находится в списке ожидания. Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗",
			text.String())
	})

	t.Run("TmplCanceledWithSingle", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplCanceledWithSingle,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> отменил вашу регистрацию. Я вернул тебя в список ищущих пару 🤗",
			text.String())
	})

	t.Run("TmplCanceledByPartner", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplCanceledByPartner,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> отменил вашу регистрацию.",
			text.String())
	})

	t.Run("TmplAutoPairPartnerFound", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplAutoPairPartnerFound,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\nЯ подобрал тебе в пару <a href=\"tg://user?id=1\">Test Partner</a> 👌",
			text.String())
	})

	t.Run("TmplAutoPairPartnerFound_waitlist", func(t *testing.T) {
		payload := testPayload
		payload.WaitList = true
		n := &models.Notification{
			TmplCode: models.TmplAutoPairPartnerFound,
			Payload:  payload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\nЯ подобрал тебе в пару <a href=\"tg://user?id=1\">Test Partner</a> 👌"+
				"\n\nВаша пара находится в списке ожидания. Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗",
			text.String())
	})

	t.Run("TmplAutoPairPartnerChanged", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplAutoPairPartnerChanged,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> отменил вашу регистрацию. \nЯ записал тебя вместе с <a href=\"https://t.me/new_partner\">Another</a> 👌",
			text.String())
	})

	t.Run("TmplAutoPairPartnerChanged_waitlist", func(t *testing.T) {
		payload := testPayload
		payload.WaitList = true
		n := &models.Notification{
			TmplCode: models.TmplAutoPairPartnerChanged,
			Payload:  payload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> отменил вашу регистрацию. \nЯ записал тебя вместе с <a href=\"https://t.me/new_partner\">Another</a> 👌"+
				"\n\nВаша пара находится в списке ожидания. Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗",
			text.String())
	})

	t.Run("TmplCoupleWaitListLeft", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplCoupleWaitListLeft,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\nВы вместе с <a href=\"tg://user?id=1\">Test Partner</a> вышли из списка ожидания 🎉\n\nЕсли планы изменились, и вы не сможете принять участие, пожалуйста, отмените вашу регистрацию.",
			text.String())
	})

	t.Run("TmplEventLimitIncreased", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplEventLimitIncreased,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=100\">Test Owner</a> увеличил лимит пар и вы вместе с <a href=\"tg://user?id=1\">Test Partner</a> вышли из списка ожидания 🎉\n\nЕсли планы изменились, и вы не сможете принять участие, пожалуйста, отмените вашу регистрацию.",
			text.String())
	})

	t.Run("TmplEventLimitDecreased", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplEventLimitDecreased,
			Payload:  testPayload,
		}
		text, err := notifyTextBuilder(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=100\">Test Owner</a> уменьшил лимит пар и вы вместе с <a href=\"tg://user?id=1\">Test Partner</a> теперь в списке ожидания.\n\nЕсли кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗",
			text.String())
	})
}

func TestNotifyTextNewEventChatName(t *testing.T) {
	tests := []struct {
		name string
		chat models.Chat
		want string
	}{
		{
			name: "title",
			chat: models.Chat{Title: "Dance & Practice"},
			want: "🔔 Новая запись в Dance &amp; Practice.",
		},
		{
			name: "username",
			chat: models.Chat{Username: "dance_practice"},
			want: "🔔 Новая запись в @dance_practice.",
		},
		{
			name: "unnamed chat",
			chat: models.Chat{},
			want: "🔔 Новая запись в чате.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &models.Notification{
				TmplCode: models.TmplNewEvent,
				Payload: models.NotificationPayload{Event: &models.Event{Post: &models.Post{
					Chat: &tt.chat,
				}}},
			}
			text, err := notifyTextBuilder(n)
			require.NoError(t, err)
			require.Equal(t, tt.want, text.String())
		})
	}
}

func Test_btnNewEventNotification(t *testing.T) {
	event := &models.Event{
		ID: "event_id",
		Post: &models.Post{
			Chat:          &models.Chat{ID: -1001234567890, Type: models.ChatSuper},
			ChatMessageID: 42,
		},
	}
	rm := btnNewEventNotification(event)
	require.Len(t, rm.InlineKeyboard, 2)
	require.Len(t, rm.InlineKeyboard[0], 1)
	require.Equal(t, locale.BtnChatLink, rm.InlineKeyboard[0][0].Text)
	require.Equal(t, "https://t.me/c/1234567890/42", rm.InlineKeyboard[0][0].URL)
	require.Equal(t, locale.BtnUnsubscribe, rm.InlineKeyboard[1][0].Text)
	require.Contains(t, rm.InlineKeyboard[1][0].Data, "event_id")
}

func Test_notifyTextEscapesDancerAndProfileNames(t *testing.T) {
	payload := testPayload
	payload.Event.Caption = "Event <Tag> & Co"
	payload.Event.Owner = models.Profile{
		ID:        100,
		FirstName: "Организатор <Main>",
		LastName:  "& Co",
	}
	payload.Partner = &models.Dancer{
		Profile: &models.Profile{
			ID:        1,
			FirstName: "Партнер",
		},
		FullName: "Партнер <One> & Two",
	}
	n := &models.Notification{
		TmplCode: models.TmplEventLimitDecreased,
		Payload:  payload,
	}

	text, err := notifyTextBuilder(n)
	require.NoError(t, err)
	got := text.String()

	assert.Contains(t, got, "Event &lt;Tag&gt; &amp; Co")
	assert.Contains(t, got, `<a href="tg://user?id=100">Организатор &lt;Main&gt; &amp; Co</a>`)
	assert.Contains(t, got, `<a href="tg://user?id=1">Партнер &lt;One&gt; &amp; Two</a>`)
	assert.NotContains(t, got, "Event <Tag>")
	assert.NotContains(t, got, "Организатор <Main>")
	assert.NotContains(t, got, "Партнер <One>")
}

func Test_chatLink(t *testing.T) {
	tests := []struct {
		name  string
		event *models.Event
		want  string
		ok    bool
	}{
		{
			name: "supergroup post",
			event: &models.Event{
				Post: &models.Post{
					Chat:          &models.Chat{ID: -1001234567890, Type: models.ChatSuper},
					ChatMessageID: 42,
				},
			},
			want: "https://t.me/c/1234567890/42",
			ok:   true,
		},
		{
			name: "channel post",
			event: &models.Event{
				Post: &models.Post{
					Chat:          &models.Chat{ID: -1001234567890, Type: models.ChatChannel},
					ChatMessageID: 42,
				},
			},
			want: "https://t.me/c/1234567890/42",
			ok:   true,
		},
		{
			name: "public channel post",
			event: &models.Event{
				Post: &models.Post{
					Chat: &models.Chat{
						ID: -1001234567890, Type: models.ChatChannel, Username: "dance_channel",
					},
					ChatMessageID: 42,
				},
			},
			want: "https://t.me/dance_channel/42",
			ok:   true,
		},
		{
			name:  "nil event",
			event: nil,
			ok:    false,
		},
		{
			name:  "nil post",
			event: &models.Event{},
			ok:    false,
		},
		{
			name: "removed event",
			event: &models.Event{
				Removed: true,
				Post: &models.Post{
					Chat:          &models.Chat{ID: -1001234567890, Type: models.ChatSuper},
					ChatMessageID: 42,
				},
			},
			ok: false,
		},
		{
			name: "private chat",
			event: &models.Event{
				Post: &models.Post{
					Chat:          &models.Chat{ID: 100, Type: models.ChatPrivate},
					ChatMessageID: 42,
				},
			},
			ok: false,
		},
		{
			name: "missing message id",
			event: &models.Event{
				Post: &models.Post{
					Chat: &models.Chat{ID: -1001234567890, Type: models.ChatSuper},
				},
			},
			ok: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := chatLink(tt.event)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.want, got)
		})
	}
}

var testPayload = models.NotificationPayload{
	Event: &models.Event{
		Caption: "Test Event",
		Owner:   models.Profile{ID: 100, FirstName: "Test", LastName: "Owner"},
		Post:    &models.Post{Chat: &models.Chat{Title: "Test Chat"}},
	},
	Partner: &models.Dancer{
		Profile: &models.Profile{
			ID:        1,
			FirstName: "Test",
			LastName:  "Partner",
		},
		FullName: "Test Partner",
	},
	NewPartner: &models.Dancer{
		Profile: &models.Profile{
			ID:        2,
			FirstName: "Another",
			Username:  "new_partner",
		},
		FullName: "Another",
	},
}
