package handlers

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// CbEventSignup handles signup callback buttons.
// Adds post to the event, re-renders event post, redirects user to signup deeplink
func (h *Handlers) CbEventSignup(c tele.Context) error {
	h.log.Info("[handlers] event signup callback received", telelog.Attr(c))

	if len(c.Args()) < 2 {
		h.log.Error("[handlers] event signup callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	// Add post to the event and re-render it
	eventID := c.Args()[0]
	inlineMessageID := c.Callback().MessageID
	event, post, err := h.eventService.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] event signup callback: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event signup callback: event post added",
		"event", event.LogValue(),
		"post", post.LogValue(),
		telelog.Trace(c))
	role := models.Role(c.Args()[1])
	return c.Respond(&tele.CallbackResponse{URL: views.EventSignupURL(eventID, role)})
}

// CbEventClosed - handles post closed for all callback button.
func (h *Handlers) CbEventClosed(c tele.Context) error {
	h.log.Info("[handlers] event closed callback received", telelog.Attr(c))
	return c.RespondText(locale.ResultEventClosed)
}

// EventQuery - handles inline query.
// If query is not empty creates draft event.
func (h *Handlers) EventQuery(c tele.Context) error {
	if c.Query().Text == "" {
		return views.EventCreateAnswerEmpty(c, h.cfg.QueryThumbUrl)
	}

	u := h.userGet(c)
	event, err := h.eventService.Create(h.ctx(c), c.Query().Text, u.Profile, u.Settings.Event)
	if err != nil {
		h.log.Error("[handlers] failed to create event: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event created", "event", event.LogValue(), telelog.Trace(c))
	return views.EventCreateAnswer(c, event.ID, h.cfg.QueryThumbUrl)
}

// EventInlineResult handles chosen inline result.
// Adds post to the event and re-renders event post.
func (h *Handlers) EventInlineResult(c tele.Context) error {
	h.log.Info("[handlers] event inline result received", telelog.Attr(c))
	eventID := c.InlineResult().ResultID
	inlineMessageID := c.InlineResult().MessageID

	// Skip if no message ID or empty query
	if inlineMessageID == "" || c.InlineResult().Query == "" {
		return nil
	}

	// Add post to the event and re-render it
	event, post, err := h.eventService.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] event inline result: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			"inline_message_id", inlineMessageID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event inline result: event post added",
		"event", event.LogValue(),
		"post", post.LogValue(),
		telelog.Trace(c))
	return nil
}
