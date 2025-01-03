package views

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
)

func Test_notifyText(t *testing.T) {
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

var testPayload = models.NotificationPayload{
	Event: &models.Event{
		Caption: "Test Event",
		Owner:   models.Profile{ID: 100, FirstName: "Test", LastName: "Owner"},
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
