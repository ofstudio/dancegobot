package handlers

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// Start - handle /start command.
// If the command has a payload, handle it as a Deeplink.
func (h *Handlers) Start(c tele.Context) error {
	h.log.Info("[handlers] /start received", "payload", c.Message().Payload, telelog.Attr(c))
	u := h.userGet(c)

	if c.Message().Payload != "" {
		u.Session = models.Session{}
		h.userSessionUpdate(c, u)
		dl, err := telegram.DeeplinkParsePayload(c.Message().Payload)
		if err != nil {
			h.log.Error("[handlers] /start: failed to parse deeplink payload: "+err.Error(), telelog.Trace(c))
			return h.sendErr(c, locale.ErrStartPayload)
		}
		switch dl.Action {
		case models.SessionSignup:
			return h.signupScene(c, dl.EventID, dl.Role)
		default:
			return h.sendErr(c, locale.ErrStartPayload)
		}
	}

	// Reset active signup scene before showing the start message.
	if u.Session.Action != "" {
		h.userSessionResetSignup(c)
	}
	return views.SendStart(c)
}
