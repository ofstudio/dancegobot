package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
)

var testPayload = models.NotificationPayload{
	Event: &models.Event{
		Caption: "Test Event",
	},
	Partner: &models.Dancer{
		Profile: &models.Profile{
			ID: 1,
		},
		FullName: "Test Partner",
	},
	NewPartner: &models.Dancer{
		Profile: &models.Profile{
			ID:       2,
			Username: "new_partner",
		},
		FullName: "New Partner",
	},
}

func TestExecTmpl(t *testing.T) {
	t.Run("TmplRegisteredWithSingle", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplRegisteredWithSingle,
			Payload:  testPayload,
		}
		text, err := notifyText(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> зарегистрировался с тобой в паре! 🎉",
			text.String())
	})

	t.Run("TmplCanceledWithSingle", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplCanceledWithSingle,
			Payload:  testPayload,
		}
		text, err := notifyText(n)
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
		text, err := notifyText(n)
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
		text, err := notifyText(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\nЯ подобрал тебе в пару <a href=\"tg://user?id=1\">Test Partner</a> 👌",
			text.String())
	})

	t.Run("TmplAutoPairPartnerChanged", func(t *testing.T) {
		n := &models.Notification{
			TmplCode: models.TmplAutoPairPartnerChanged,
			Payload:  testPayload,
		}
		text, err := notifyText(n)
		require.NoError(t, err)
		assert.Equal(t,
			"🔔 Test Event\n\n<a href=\"tg://user?id=1\">Test Partner</a> отменил вашу регистрацию. \nЯ записал тебя вместе с <a href=\"https://t.me/new_partner\">New Partner</a> 👌",
			text.String())
	})
}
