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

// EventSettingsScene handles event settings callback button.
func (h *Handlers) EventSettingsScene(c tele.Context) error {
	h.log.Info("[handlers] event settings scene received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings scene: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset, _ := strconv.Atoi(c.Args()[1])

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings scene: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	if !h.eventService.CanManage(event, u.Profile) {
		h.log.Error("[handlers] event settings scene: user can't manage the event",
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	_ = c.Respond()
	return views.EventSettingsScene(c, event, offset)
}

// CbEventSettingsBack - handles event settings scene back callback button.
func (h *Handlers) CbEventSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] event settings back callback received", telelog.Attr(c))
	if len(c.Args()) < 2 {
		h.log.Error("[handlers] event settings back callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}
	offset, _ := strconv.Atoi(c.Args()[0])
	_ = c.Respond()
	return h.myScene(c, offset)
}

// CbEventSettingsToggles - handles event settings scene callback buttons toggles.
func (h *Handlers) CbEventSettingsToggles(c tele.Context) error {
	h.log.Info("[handlers] event settings toggle callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings toggle callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	u := h.userGet(c)
	eventID := c.Args()[0]
	offset, _ := strconv.Atoi(c.Args()[1])

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings toggle callback: "+err.Error(),
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

	if event, err = h.eventService.SettingsUpdate(h.ctx(c), event.ID, u.Profile, event.Settings); err != nil {
		h.log.Error("[handlers] event settings toggle callback: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] event settings updated",
		"event", event.LogValue(),
		"settings", event.Settings.LogValue(),
		telelog.Trace(c))

	_ = c.Respond()
	return views.EventSettingsScene(c, event, offset)
}

// CbEventSettingsLimitScene handles event settings limit scene.
func (h *Handlers) CbEventSettingsLimitScene(c tele.Context) error {
	h.log.Info("[handlers] event settings limit scene callback received", telelog.Attr(c))
	if len(c.Args()) < 4 {
		h.log.Error("[handlers] event settings limit scene callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	eventID := c.Args()[0]
	page, _ := strconv.Atoi(c.Args()[1])
	offset := c.Args()[2]

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings limit scene callback: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	_ = c.Respond()
	return views.EventSettingsLimitScene(c, event.ID, page, offset)
}

// CbEventSettingsLimitNum handles event settings limit number callback buttons
// in event settings limit scene.
func (h *Handlers) CbEventSettingsLimitNum(c tele.Context) error {
	h.log.Info("[handlers] event settings limit number callback received", telelog.Attr(c))
	if len(c.Args()) < 3 {
		h.log.Error("[handlers] event settings limit number callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	eventID := c.Args()[0]
	limit, _ := strconv.Atoi(c.Args()[1])
	offset, _ := strconv.Atoi(c.Args()[2])

	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] event settings limit number callback: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	oldLimit := event.Settings.Limit
	event.Settings.Limit = limit
	if event, err = h.eventService.SettingsUpdate(h.ctx(c), event.ID, h.userGet(c).Profile, event.Settings); err != nil {
		h.log.Error("[handlers] event settings limit number callback: "+err.Error(),
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
		views.EventSettingsScene(c, event, offset),
		h.limitChangedNotice(c, event, oldLimit),
	)
}

// CbLimitChangedNotify handles the limit changed notify callback.
func (h *Handlers) CbLimitChangedNotify(c tele.Context) error {
	h.log.Info("[handlers] limit changed notify callback received", telelog.Attr(c))

	if len(c.Args()) < 2 {
		h.log.Error("[handlers] limit changed notify callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	eventID := c.Args()[0]
	if eventID == "" {
		h.log.Error("[handlers] limit changed notify callback: empty event ID", telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	// Remove inline keyboard
	_ = c.Respond()
	if err := views.RemoveInlineKeyboard(c); err != nil {
		h.log.Error("[handlers] limit changed notify callback: "+err.Error(),
			telelog.Trace(c))
	}

	// Check for the event ID in the user session
	u := h.userGet(c)
	if u.Session.EventID != eventID {
		h.log.Info("[handlers] limit changed notify callback: event ID mismatch",
			"event_id", eventID,
			"session_event_id", u.Session.EventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.LimitChangedCantNotify)
	}

	// Check for the affected couples in the user session
	if u.Session.AffectedCouples == nil {
		h.log.Info("[handlers] limit changed notify callback: no affected couples in the session",
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.LimitChangedCantNotify)
	}

	// Reset affected couples in the user session
	affected := *u.Session.AffectedCouples
	h.userSessionUpdateAffected(c, "", nil)

	// Send notifications to the affected couples
	if err := h.eventService.LimitChangeNotifyAffected(h.ctx(c), eventID, affected); err != nil {
		h.log.Error("[handlers] limit changed notify callback: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] limit changed notify callback: notifications sent",
		"event_id", eventID, telelog.Trace(c))
	return views.SendLimitChangedNotified(c)
}

// CbLimitChangedSkip handles the limit changed skip notification callback.
func (h *Handlers) CbLimitChangedSkip(c tele.Context) error {
	h.log.Info("[handlers] limit changed skip notification callback received", telelog.Attr(c))
	_ = c.Respond()

	// Reset affected couples in the user session
	h.userSessionUpdateAffected(c, "", nil)

	return views.RemoveInlineKeyboard(c)
}

// limitChangedNotice sends a notice about the limit change.
func (h *Handlers) limitChangedNotice(c tele.Context, event *models.Event, oldLimit int) error {
	// Skip if the limit hasn't changed
	if event.Settings.Limit == oldLimit {
		return nil
	}

	// Get affected couples
	affected := h.eventService.LimitChangeGetAffected(event, oldLimit)

	// Skip if no couples affected
	if len(affected.Couples) == 0 {
		return nil
	}

	// Update user session with the affected couples
	h.userSessionUpdateAffected(c, event.ID, &affected)

	return views.SendLimitChanged(c, event, affected)
}

// userSessionUpdateAffected updates the user session with the affected couples.
func (h *Handlers) userSessionUpdateAffected(c tele.Context, eventID string, affected *models.AffectedCouples) {
	u := h.userGet(c)
	u.Session.EventID = eventID
	u.Session.AffectedCouples = affected
	h.userSessionUpdate(c, u)
}
