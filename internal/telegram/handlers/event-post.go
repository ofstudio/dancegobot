package handlers

import (
	"regexp"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// EventQuery - handles inline query.
// If query is not empty creates draft event.
func (h *Handlers) EventQuery(c tele.Context) error {
	text := strings.TrimSpace(c.Query().Text)
	if text == "" {
		return views.EventPostAnswerEmpty(c, h.cfg.QueryThumbUrl)
	}

	u := h.userGet(c)
	text, settings := h.eventQueryParse(text, u.Settings.Event)

	event, err := h.eventService.Create(h.ctx(c), text, u.Profile, settings)
	if err != nil {
		h.log.Error("[handlers] event query: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	h.log.Info("[handlers] event query: event created", "event", event.LogValue(), telelog.Trace(c))
	return views.EventPostAnswer(c, event, h.cfg.QueryThumbUrl)
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
	event, err := h.eventService.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] event inline result: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			"inline_message_id", inlineMessageID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event inline result: event post added",
		"event", event.LogValue(),
		"post", event.Post.LogValue(),
		telelog.Trace(c))
	return nil
}

// CbEventSignup handles signup callback buttons.
// Adds post to the event, re-renders event post, redirects user to signup deeplink
func (h *Handlers) CbEventSignup(c tele.Context) error {
	h.log.Info("[handlers] event signup callback received", telelog.Attr(c))

	if len(c.Args()) < 2 {
		h.log.Error("[handlers] event signup callback: not enough arguments",
			"args", c.Args(),
			telelog.Trace(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	// Add post to the event and re-render it
	eventID := c.Args()[0]
	inlineMessageID := c.Callback().MessageID
	event, err := h.eventService.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] event signup callback: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.respondErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event signup callback: event post added",
		"event", event.LogValue(),
		"post", event.Post.LogValue(),
		telelog.Trace(c))
	role := models.Role(c.Args()[1])
	return c.Respond(&tele.CallbackResponse{URL: views.EventSignupURL(eventID, role)})
}

// CbEventClosed - handles post closed for all callback button.
func (h *Handlers) CbEventClosed(c tele.Context) error {
	h.log.Info("[handlers] event closed callback received", telelog.Attr(c))
	return c.RespondText(locale.ResultEventClosed)
}

// eventQueryParse parses the inline query text.
// Takes inline query text and default event settings.
// Finds settings shortcuts in inline query text and returns
// the rest of the text and overridden event settings.
func (h *Handlers) eventQueryParse(text string, settings models.EventSettings) (string, models.EventSettings) {
	text, settings = h.eventQueryParseLimit(text, settings)
	return text, settings
}

// reEventLimit is a regexp pattern for the event limit in the inline query text.
// The limit is a standalone token from /1 to /99, e.g. " /5".
// Regexp explanation:
//
//	(whitespace)/(1..99)(whitespace or end)
var reEventLimit = regexp.MustCompile(`\s/([1-9]\d?)(?:\s|$)`)

// eventQueryParseLimit parses the inline query text for the event limit.
func (h *Handlers) eventQueryParseLimit(text string, settings models.EventSettings) (string, models.EventSettings) {
	matches := reEventLimit.FindStringSubmatchIndex(text)
	if len(matches) < 4 {
		return text, settings
	}
	limit, _ := strconv.Atoi(text[matches[2]:matches[3]])
	settings.Limit = limit
	text = strings.TrimSpace(text[:matches[0]] + " " + text[matches[1]:])
	return text, settings
}
