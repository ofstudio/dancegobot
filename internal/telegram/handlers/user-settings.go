package handlers

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// CbUserSettingsAutoPair - toggles auto pair user setting.
func (h *Handlers) CbUserSettingsAutoPair(c tele.Context) error {
	h.log.Info("[handlers] user settings auto pair callback received", telelog.Attr(c))
	u := h.userGet(c)
	u.Settings.Event.AutoPairing = !u.Settings.Event.AutoPairing
	h.userUpdateSettings(c, u)
	_ = c.Respond()
	text, rm := views.UserSettingsScene(&u.Settings)
	return c.Edit(text, rm, tele.ModeHTML)
}

// CbUserSettingsHelp - sends settings help message.
func (h *Handlers) CbUserSettingsHelp(c tele.Context) error {
	h.log.Info("[handlers] user settings help callback received", telelog.Attr(c))
	_ = c.Respond()
	text, rm := views.UserSettingsHelp()
	return c.Edit(text, rm, tele.ModeHTML, tele.NoPreview)
}

// CbUserSettingsBack - sends settings scene message.
func (h *Handlers) CbUserSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] user settings back callback received", telelog.Attr(c))
	u := h.userGet(c)
	_ = c.Respond()
	text, rm := views.UserSettingsScene(&u.Settings)
	return c.Edit(text, rm, tele.ModeHTML)
}

// UserSettings - handles /settings command.
func (h *Handlers) UserSettings(c tele.Context) error {
	h.log.Info("[handlers] /settings received", telelog.Attr(c))
	u := h.userGet(c)
	text, rm := views.UserSettingsScene(&u.Settings)
	return c.Send(text, rm, tele.ModeHTML)
}
