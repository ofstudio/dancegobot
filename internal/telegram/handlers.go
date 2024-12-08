package telegram

import (
	"context"
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

var nowFn = func() time.Time {
	return time.Now().UTC()
}

type Handlers struct {
	cfg    config.Settings
	events EventService
	users  UserService
	log    *slog.Logger
}

func NewHandlers(cfg config.Settings, es EventService, us UserService) *Handlers {
	return &Handlers{
		cfg:    cfg,
		events: es,
		users:  us,
		log:    noplog.Logger(),
	}
}

func (h *Handlers) WithLogger(l *slog.Logger) *Handlers {
	h.log = l
	return h
}

// ctx returns the context from the telebot context.
// If the context is not set, it returns a new context.Background().
func (h *Handlers) ctx(c tele.Context) context.Context {
	ctx, ok := c.Get("ctx").(context.Context)
	if !ok {
		ctx = context.Background()
	}
	return ctx
}

// sessionSet saves the user session.
func (h *Handlers) sessionSet(c tele.Context, session models.Session) {
	if err := h.users.SessionUpsert(h.ctx(c), &models.User{
		Profile: models.NewProfile(*c.Sender()),
		Session: session,
	}); err != nil {
		h.log.Error("[handlers] failed to set session: "+err.Error(), telelog.Trace(c))
	}
}

// sessionReset resets the user session.
func (h *Handlers) sessionReset(c tele.Context) {
	h.sessionSet(c, models.Session{})
}

// sessionGet returns the user session.
// If the session is not found or an error occurred, returns an empty session.
func (h *Handlers) sessionGet(c tele.Context) models.Session {
	user, err := h.users.Get(h.ctx(c), models.NewProfile(*c.Sender()))
	if err != nil {
		h.log.Error("[handlers] failed to get session: "+err.Error(), telelog.Trace(c))
		return models.Session{}
	}
	return user.Session
}

// Start - handle /start command.
// If the command has a payload, handle it as a Deeplink.
func (h *Handlers) Start(c tele.Context) error {
	h.log.Info("[handlers] /start received", "payload", c.Message().Payload, telelog.Attr(c))

	if c.Message().Payload != "" {
		h.sessionReset(c)
		action, params, err := deeplinkParsePayload(c.Message().Payload)
		if err != nil {
			h.log.Error("[handlers] /start: failed to parse deeplink payload: "+err.Error(), telelog.Trace(c))
			return h.sendErr(c, locale.ErrStartPayload)
		}
		switch action {
		case models.SessionSignup:
			return h.signupScene(c, params[0], models.Role(params[1]))
		default:
			return h.sendErr(c, locale.ErrStartPayload)
		}
	}

	// Send start message only if user session is empty
	// due to some Telegram clients (ie iOS, late 2024)
	// can "double" /start messages on very first interaction with the bot
	session := h.sessionGet(c)
	if session.Action == "" {
		return sendStart(c)
	}
	return nil
}

// Query - handles inline query.
// If query is not empty creates draft event.
func (h *Handlers) Query(c tele.Context) error {
	if c.Query().Text == "" {
		return answerQueryEmpty(c, h.cfg.QueryThumbUrl)
	}
	eventID := h.events.NewID()

	event := models.Event{
		ID:        eventID,
		Caption:   c.Query().Text,
		Owner:     models.NewProfile(*c.Sender()),
		CreatedAt: nowFn(),
	}
	if err := h.events.Create(h.ctx(c), &event); err != nil {
		h.log.Error("[handlers] failed to create event: "+err.Error(), "event", event, telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event created", "event", event, telelog.Trace(c))
	return answerQuery(c, eventID, h.cfg.QueryThumbUrl)
}

// InlineResult handles chosen inline result.
// Adds post to the event and re-renders event post.
func (h *Handlers) InlineResult(c tele.Context) error {
	h.log.Info("[handlers] chosen_inline_result received", telelog.Attr(c))
	eventID := c.InlineResult().ResultID
	inlineMessageID := c.InlineResult().MessageID

	// Skip if no message ID or empty query
	if inlineMessageID == "" || c.InlineResult().Query == "" {
		return nil
	}

	// Add post to the event and re-render it
	upd, err := h.events.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] chosen_inline_result: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] chosen_inline_result: event post added", "", upd, telelog.Trace(c))
	return nil
}

// CbSignup handles signup callback buttons.
// Adds post to the event, re-renders event post, redirects user to signup deeplink
func (h *Handlers) CbSignup(c tele.Context) error {
	h.log.Info("[handlers] signup callback received", telelog.Attr(c))

	if len(c.Args()) < 2 {
		h.log.Error("[handlers] signup callback: not enough arguments",
			"args", c.Args(),
			telelog.Attr(c))
		return c.RespondAlert(locale.ErrSomethingWrong)
	}

	// Add post to the event and re-render it
	eventID := c.Args()[0]
	inlineMessageID := c.Callback().MessageID
	upd, err := h.events.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] signup callback: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] signup callback: event post added", "", upd, telelog.Trace(c))
	role := c.Args()[1]
	return c.Respond(&tele.CallbackResponse{URL: deeplink(models.SessionSignup, eventID, role)})
}

// UserShared - handles the user shared event.
func (h *Handlers) UserShared(c tele.Context) error {
	h.log.Info("[handlers] users_shared received", telelog.Attr(c))
	s := h.sessionGet(c)
	if s.Action != models.SessionSignup {
		h.log.Error("[handlers] unexpected user_shared", telelog.Trace(c))
		return nil
	}

	userShared := c.Message().UserShared.Users[0]
	other := models.Profile{
		ID:        userShared.UserID,
		FirstName: userShared.FirstName,
		LastName:  userShared.LastName,
		Username:  userShared.Username,
	}

	return h.coupleAdd(c, s.EventID, s.Role, &other)

}

// Partner - handles the /partner command (legacy).
// This is to provide familiar user experience with github.com/Tayrinn/CoopDance bot.
// The payload will be treated as a text message.
func (h *Handlers) Partner(c tele.Context) error {
	h.log.Warn("[handlers] /partner received: payload will be treated as text message", telelog.Attr(c))
	if c.Message().Payload == "" {
		return nil
	}
	c.Message().Text = c.Message().Payload
	c.Message().Payload = ""
	return h.Text(c)
}

// Text - handles text messages.
func (h *Handlers) Text(c tele.Context) error {
	h.log.Info("[handlers] text message received", "text", c.Text(), telelog.Attr(c))

	s := h.sessionGet(c)
	text := c.Text()

	switch {
	case s.Action != models.SessionSignup:
		h.log.Info("[handlers] unexpected text", telelog.Trace(c))
		return nil // todo maybe some help message or random joke or facts?
	case text == locale.BtnClose:
		h.sessionReset(c)
		return sendCloseOK(c)
	case text == locale.BtnRemove:
		return h.dancerRemove(c, s.EventID)
	case text == locale.BtnSingle[s.Role]:
		return h.singleAdd(c, s.EventID, s.Role)
	case isSingleCaption(text):
		for _, single := range s.Singles {
			if single.Caption == text {
				return h.coupleAdd(c, s.EventID, s.Role, &single.Profile)
			}
		}
		return h.sendErr(c, locale.ErrSingleNotFound)
	case len(text) > h.cfg.DancerNameMaxLen:
		return h.sendErr(c, locale.ErrDancerNameTooLong)
	default:
		return h.coupleAdd(c, s.EventID, s.Role, text)
	}
}

// signupScene returns the signup scene for the user.
func (h *Handlers) signupScene(c tele.Context, eventID string, role models.Role) error {
	event, err := h.events.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] signup scene: failed to get event: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	profile := models.NewProfile(*c.Sender())
	dancer := h.events.DancerGet(event, &profile, role)
	role = dancer.Role

	// if the dancer can signup, update the session
	var singles []models.SessionSingle
	if dancer.Status.SignupAvailable() || dancer.Status.SignedUp() {
		singles = fmtSingles(event.Singles, role.Opposite())
		h.sessionSet(c, models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    dancer.Role,
			Singles: singles,
		})
	} else {
		// otherwise, reset the session
		h.sessionReset(c)
	}

	h.log.Info("[handlers] signup scene", "event_id", eventID, "dancer", dancer, telelog.Trace(c))

	return sendSignup(c, dancer, singles)
}

// coupleAdd handles the couple signup action
func (h *Handlers) coupleAdd(c tele.Context, eventID string, role models.Role, other any) error {
	profile := models.NewProfile(*c.Sender())

	upd, err := h.events.CoupleAdd(h.ctx(c), eventID, &profile, role, other)
	if err != nil {
		h.log.Error("[handlers] failed to add couple: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] couple added", "", upd, telelog.Trace(c))

	// if the result is retryable, update the session
	var singles []models.SessionSingle
	if upd.Result.Retryable() {
		singles = fmtSingles(upd.Event.Singles, role.Opposite())
		h.sessionSet(c, models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    role,
			Singles: singles,
		})
	} else {
		// otherwise, reset the session
		h.sessionReset(c)
	}

	return sendResult(c, upd, singles)
}

// singleAdd handles the single signup action
func (h *Handlers) singleAdd(c tele.Context, eventID string, role models.Role) error {
	profile := models.NewProfile(*c.Sender())

	upd, err := h.events.SingleAdd(h.ctx(c), eventID, &profile, role)
	if err != nil {
		h.log.Error("[handlers] failed to add single: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] single added", "", upd, telelog.Trace(c))

	// if the result is retryable, update the session
	var singles []models.SessionSingle
	if upd.Result.Retryable() {
		singles = fmtSingles(upd.Event.Singles, role.Opposite())
		h.sessionSet(c, models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    role,
			Singles: singles,
		})
	} else {
		// otherwise, reset the session
		h.sessionReset(c)
	}

	return sendResult(c, upd, singles)
}

// dancerRemove handles the dancer remove action
func (h *Handlers) dancerRemove(c tele.Context, eventID string) error {
	profile := models.NewProfile(*c.Sender())

	upd, err := h.events.DancerRemove(h.ctx(c), eventID, &profile)
	if err != nil {
		h.log.Error("[handlers] failed to remove dancer: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] dancer removed", "", upd, telelog.Trace(c))

	h.sessionReset(c)

	return sendResult(c, upd, nil)
}

// sendErr sends an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) sendErr(c tele.Context, msg string) error {
	h.sessionReset(c)
	return c.Send(msg, tele.RemoveKeyboard)
}
