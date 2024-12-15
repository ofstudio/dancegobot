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

// userGet returns [models.User] from [tele.Context].
// If the user is not found in the context, logs an error
// and returns user with current profile and empty session and settings.
func (h *Handlers) userGet(c tele.Context) *models.User {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		h.log.Error("[handlers] user not found in context. This can cause unexpected behavior", telelog.Trace(c))
		user = &models.User{
			Profile: models.NewProfile(*c.Sender()),
		}
	}
	return user
}

// userUpsert saves the user.
func (h *Handlers) userUpsert(c tele.Context, user *models.User) {
	if err := h.users.Upsert(h.ctx(c), user); err != nil {
		h.log.Error("[handlers] failed to upsert user: "+err.Error(), telelog.Trace(c))
	}
}

// Start - handle /start command.
// If the command has a payload, handle it as a Deeplink.
func (h *Handlers) Start(c tele.Context) error {
	h.log.Info("[handlers] /start received", "payload", c.Message().Payload, telelog.Attr(c))
	u := h.userGet(c)

	if c.Message().Payload != "" {
		u.Session = models.Session{}
		h.userUpsert(c, u)
		dl, err := DeeplinkParsePayload(c.Message().Payload)
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

	// Send start message only if user session is empty
	// due to some Telegram clients (ie iOS, late 2024)
	// can "double" /start messages on very first interaction with the bot
	if u.Session.Action == "" {
		return sendStart(c)
	}
	return nil
}

// Settings - handles /settings command.
func (h *Handlers) Settings(c tele.Context) error {
	h.log.Info("[handlers] /settings received", telelog.Attr(c))
	u := h.userGet(c)
	text, rm := msgSettingsScene(&u.Settings)
	return c.Send(text, rm, tele.ModeHTML)
}

// Query - handles inline query.
// If query is not empty creates draft event.
func (h *Handlers) Query(c tele.Context) error {
	u := h.userGet(c)
	if c.Query().Text == "" {
		return answerQueryEmpty(c, h.cfg.QueryThumbUrl)
	}
	eventID := h.events.NewID()

	event := models.Event{
		ID:      eventID,
		Caption: c.Query().Text,
		Settings: models.EventSettings{
			DisableChooseSingle: u.Settings.Events.DisableChooseSingle,
		},
		Owner:     u.Profile,
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

// CbSettingsAutoPair - toggles auto pair user setting.
func (h *Handlers) CbSettingsAutoPair(c tele.Context) error {
	h.log.Info("[handlers] settings_auto_pair callback received", telelog.Attr(c))
	u := h.userGet(c)
	u.Settings.Events.AutoPairing = !u.Settings.Events.AutoPairing
	h.userUpsert(c, u)
	_ = c.Respond()
	text, rm := msgSettingsScene(&u.Settings)
	return c.Edit(text, rm, tele.ModeHTML)
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
	dl := Deeplink{Action: models.SessionSignup, EventID: eventID, Role: models.Role(role)}
	return c.Respond(&tele.CallbackResponse{URL: dl.String()})
}

// UserShared - handles the user shared event.
func (h *Handlers) UserShared(c tele.Context) error {
	h.log.Info("[handlers] users_shared received", telelog.Attr(c))
	u := h.userGet(c)
	if u.Session.Action != models.SessionSignup {
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

	return h.coupleAdd(c, u.Session.EventID, u.Session.Role, &other)

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

	u := h.userGet(c)
	text := c.Text()

	switch {
	case u.Session.Action != models.SessionSignup:
		h.log.Info("[handlers] unexpected text", telelog.Trace(c))
		return nil // todo maybe some help message or random joke or facts?
	case text == locale.BtnClose:
		u.Session = models.Session{}
		h.userUpsert(c, u)
		return sendCloseOK(c)
	case text == locale.BtnRemove:
		return h.dancerRemove(c, u.Session.EventID)
	case text == locale.BtnAsSingle[u.Session.Role]:
		return h.singleAdd(c, u.Session.EventID, u.Session.Role)
	case text == locale.BtnChooseSingle[u.Session.Role]:
		return h.chooseSingleScene(c)
	case isSingleCaption(text):
		for _, single := range u.Session.Singles {
			if single.Caption == text {
				return h.coupleAdd(c, u.Session.EventID, u.Session.Role, &single.Profile)
			}
		}
		return h.sendErr(c, locale.ErrSingleNotFound)
	case len(text) > h.cfg.DancerNameMaxLen:
		return h.sendErr(c, locale.ErrDancerNameTooLong)
	default:
		return h.coupleAdd(c, u.Session.EventID, u.Session.Role, text)
	}
}

// signupScene returns the signup scene for the user.
func (h *Handlers) signupScene(c tele.Context, eventID string, role models.Role) error {
	event, err := h.events.Get(h.ctx(c), eventID)
	u := h.userGet(c)
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
		u.Session = models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    dancer.Role,
			Singles: singles,
		}

	} else {
		// otherwise, reset the session
		u.Session = models.Session{}
	}
	h.userUpsert(c, u)

	h.log.Info("[handlers] signup scene", "event_id", eventID, "dancer", dancer, telelog.Trace(c))

	return sendSignupScene(c, dancer, singles)
}

// chooseSingleScene returns the scene of choosing a single partner to sign up as a couple with
func (h *Handlers) chooseSingleScene(c tele.Context) error {
	// 1. Get the user session
	u := h.userGet(c)
	if u.Session.Action != models.SessionSignup {
		h.log.Info("[handlers] choose single scene unexpected", telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	// 2. Get the event
	event, err := h.events.Get(h.ctx(c), u.Session.EventID)
	if err != nil {
		h.log.Error("[handlers] choose single scene: failed to get event: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	// 3. Get the singles for the dancer's role
	singles := fmtSingles(event.Singles, u.Session.Role.Opposite())
	if len(singles) == 0 {
		h.log.Info("[handlers] choose single scene: no singles available", telelog.Trace(c))
		return sendNoSinglesAvailable(c)
	}

	// 4. Update the session
	h.log.Info("[handlers] choose single scene", "event_id", u.Session.EventID, telelog.Trace(c))
	u.Session = models.Session{
		Action:  models.SessionSignup,
		EventID: u.Session.EventID,
		Role:    u.Session.Role,
		Singles: singles,
	}
	h.userUpsert(c, u)

	// 5. Send the scene
	return sendChooseSingleScene(c, singles)
}

// coupleAdd handles the couple signup action
func (h *Handlers) coupleAdd(c tele.Context, eventID string, role models.Role, other any) error {
	u := h.userGet(c)
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
		u.Session = models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    role,
			Singles: singles,
		}
	} else {
		// otherwise, reset the session
		u.Session = models.Session{}
	}
	h.userUpsert(c, u)
	return sendResult(c, upd, singles)
}

// singleAdd handles the single signup action
func (h *Handlers) singleAdd(c tele.Context, eventID string, role models.Role) error {
	u := h.userGet(c)
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
		u.Session = models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    role,
			Singles: singles,
		}
	} else {
		// otherwise, reset the session
		u.Session = models.Session{}
	}
	h.userUpsert(c, u)
	return sendResult(c, upd, singles)
}

// dancerRemove handles the dancer remove action
func (h *Handlers) dancerRemove(c tele.Context, eventID string) error {
	u := h.userGet(c)
	profile := models.NewProfile(*c.Sender())

	upd, err := h.events.DancerRemove(h.ctx(c), eventID, &profile)
	if err != nil {
		h.log.Error("[handlers] failed to remove dancer: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] dancer removed", "", upd, telelog.Trace(c))

	u.Session = models.Session{}
	h.userUpsert(c, u)
	return sendResult(c, upd, nil)
}

// sendErr sends an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) sendErr(c tele.Context, msg string) error {
	u := h.userGet(c)
	u.Session = models.Session{}
	h.userUpsert(c, u)
	return c.Send(msg, tele.RemoveKeyboard)
}
