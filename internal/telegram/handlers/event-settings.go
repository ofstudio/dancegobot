package handlers

import (
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/errutil"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// CbEventSettings - handles event settings callback button.
func (h *Handlers) CbEventSettings(c tele.Context) error {
	h.log.Info("[handlers] event settings callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset := c.Args()[1]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	if !h.eventService.CanManage(event, &h.userGet(c).Profile) {
		h.log.Error("[handlers] event settings callback: user can't manage the event",
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	_ = c.Respond()
	return c.Edit(views.EventSettingsMsg(event), views.BtnEventSettingsScene(event, offset),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// CbEventSettingsBack - handles event settings scene back callback button.
func (h *Handlers) CbEventSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] event_settings_back callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] event_settings_back callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])

	text, rm, err := h.myScene(c, offset)
	if err != nil {
		h.log.Error("[handlers] failed to get my scene: "+err.Error(), telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	_ = c.Respond()

	return c.Edit(text, rm, tele.ModeHTML, tele.RemoveKeyboard, tele.NoPreview)
}

// CbEventSettingsToggles - handles event settings scene callback buttons toggles.
func (h *Handlers) CbEventSettingsToggles(c tele.Context) error {
	h.log.Info("[handlers] event settings toggle callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings toggle callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset := c.Args()[1]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings toggle callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	var unique string
	if c.Callback() != nil {
		unique = c.Callback().Unique
	}

	switch unique {
	case views.BtnEventSettingsAutoPair.Unique:
		event.Settings.AutoPairing = !event.Settings.AutoPairing
	case views.BtnEventSettingsClose.Unique:
		event.Settings.Closed = !event.Settings.Closed
	default:
		h.log.Error("[handlers] event settings toggle callback: unknown unique",
			"unique", unique,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	if event, err = h.eventService.UpdateSettings(h.ctx(c), event.ID, &u.Profile, event.Settings); err != nil {
		h.log.Error("[handlers] event settings toggle callback: failed to update event settings: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] event settings updated",
		"event", event.LogValue(),
		"settings", event.Settings.LogValue(),
		telelog.Trace(c))

	_ = c.Respond()
	return c.Edit(views.EventSettingsMsg(event), views.BtnEventSettingsScene(event, offset),
		tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// CbEventSettingsLimitScene handles event settings limit scene.
func (h *Handlers) CbEventSettingsLimitScene(c tele.Context) error {
	h.log.Info("[handlers] event settings limit scene callback received", telelog.Attr(c))
	if len(c.Args()) < 4 {
		h.log.Error("[handlers] event settings limit scene callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	eventID := c.Args()[0]
	page, _ := strconv.Atoi(c.Args()[1])
	offset := c.Args()[2]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings limit scene callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	msg := views.EventSettingsMsg(event)
	rm := views.BtnEventSettingsLimitScene(eventID, page, offset)
	return c.Edit(msg, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// CbEventSettingsLimitNum handles event settings limit number callback buttons
// in event settings limit scene.
func (h *Handlers) CbEventSettingsLimitNum(c tele.Context) error {
	h.log.Info("[handlers] event settings limit number callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings limit number callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	eventID := c.Args()[0]
	limit, _ := strconv.Atoi(c.Args()[1])
	offset := c.Args()[2]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings limit number callback: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	oldLimit := event.Settings.Limit
	event.Settings.Limit = limit
	if event, err = h.eventService.UpdateSettings(h.ctx(c), event.ID, &h.userGet(c).Profile, event.Settings); err != nil {
		h.log.Error("[handlers] event settings limit number callback: failed to update event settings: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] event settings updated",
		"event", event.LogValue(),
		"settings", event.Settings.LogValue(),
		telelog.Trace(c))

	_ = c.Respond()
	return errutil.Append(
		c.Edit(views.EventSettingsMsg(event), views.BtnEventSettingsScene(event, offset),
			tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard),
		h.limitChangedPrompt(c, event, oldLimit),
	)
}

// CbLimitChangedNotify handles the limit changed notification callback.
func (h *Handlers) CbLimitChangedNotify(c tele.Context) error {
	h.log.Info("[handlers] limit changed notification callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] limit changed notification callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	eventID := c.Args()[0]
	oldLimit, err := strconv.Atoi(c.Args()[1])
	if err != nil {
		h.log.Error("[handlers] limit changed notification callback: failed to parse old limit: "+err.Error(),
			"old_limit", c.Args()[1],
			telelog.Trace(c))
	}
	_ = c.Respond()

	err = c.Edit(&tele.ReplyMarkup{}, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
	h.eventService.LimitChangeNotify(h.ctx(c), eventID, oldLimit)

	return errutil.Append(
		err,
		c.Send(locale.LimitChangedNotified, tele.RemoveKeyboard, tele.NoPreview, tele.ModeHTML),
	)
}

// CbLimitChangedSkip handles the limit changed skip notification callback.
func (h *Handlers) CbLimitChangedSkip(c tele.Context) error {
	h.log.Info("[handlers] limit changed skip notification callback received", telelog.Attr(c))
	_ = c.Respond()
	return c.Edit(&tele.ReplyMarkup{}, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// limitChangedPrompt sends a prompt message about the limit change.
func (h *Handlers) limitChangedPrompt(c tele.Context, event *models.Event, oldLimit int) error {
	// Skip if the limit hasn't changed
	if event.Settings.Limit == oldLimit {
		return nil
	}
	couples, increased, idx := h.eventService.LimitChangeAffected(event, oldLimit)
	// Skip if no couples affected
	if len(couples) == 0 {
		return nil
	}

	// Send a prompt message about affected couples
	msg := views.LimitChangedMsg(event, increased, couples, idx)
	rm := views.BtnLimitChanged(event.ID, oldLimit)
	return c.Send(msg, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}
