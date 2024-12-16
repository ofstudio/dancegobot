package telegram

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

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
		h.log.Error("[handlers] failed to upsert user: "+err.Error(),
			"profile", user.Profile.LogValue(),
			telelog.Trace(c))
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
	if c.Query().Text == "" {
		return answerQueryEmpty(c, h.cfg.QueryThumbUrl)
	}

	u := h.userGet(c)
	event, err := h.events.Create(h.ctx(c), c.Query().Text, u.Profile, u.Settings.Event)
	if err != nil {
		h.log.Error("[handlers] failed to create event: "+err.Error(), telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] event created", "event", event.LogValue(), telelog.Trace(c))
	return answerQuery(c, event.ID, h.cfg.QueryThumbUrl)
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
	event, post, err := h.events.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] chosen_inline_result: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			"inline_message_id", inlineMessageID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] chosen_inline_result: event post added",
		"event", event.LogValue(),
		"post", post.LogValue(),
		telelog.Trace(c))
	return nil
}

// CbSettingsAutoPair - toggles auto pair user setting.
func (h *Handlers) CbSettingsAutoPair(c tele.Context) error {
	h.log.Info("[handlers] settings_auto_pair callback received", telelog.Attr(c))
	u := h.userGet(c)
	u.Settings.Event.AutoPairing = !u.Settings.Event.AutoPairing
	h.userUpsert(c, u)
	_ = c.Respond()
	text, rm := msgSettingsScene(&u.Settings)
	return c.Edit(text, rm, tele.ModeHTML)
}

// CbSettingsHelp - sends settings help message.
func (h *Handlers) CbSettingsHelp(c tele.Context) error {
	h.log.Info("[handlers] settings_help callback received", telelog.Attr(c))
	_ = c.Respond()
	text, rm := msgSettingsHelp()
	return c.Edit(text, rm, tele.ModeHTML)
}

// CbSettingsBack - sends settings scene message.
func (h *Handlers) CbSettingsBack(c tele.Context) error {
	h.log.Info("[handlers] settings_back callback received", telelog.Attr(c))
	u := h.userGet(c)
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
	event, post, err := h.events.PostAdd(h.ctx(c), eventID, inlineMessageID)
	if err != nil {
		h.log.Error("[handlers] signup callback: failed add event post chat: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] signup callback: event post added",
		"event", event.LogValue(),
		"post", post.LogValue(),
		telelog.Trace(c))
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
	h.log.Info("[handlers] /partner received: payload will be treated as text message", telelog.Attr(c))
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
	u := h.userGet(c)
	event, err := h.events.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] signup scene: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	reg := h.events.RegistrationGet(event, &u.Profile, role)

	// if the dancer can register or already registered, update the session
	var singles []models.SessionSingle
	if reg.Status.CanRegister() || reg.Status.IsRegistered() {
		singles = fmtSingles(event.Singles, role.Opposite())
		u.Session = models.Session{
			Action:  models.SessionSignup,
			EventID: eventID,
			Role:    reg.Dancer.Role,
			Singles: singles,
		}
	} else {
		// otherwise, reset the session
		u.Session = models.Session{}
	}
	h.userUpsert(c, u)

	h.log.Info("[handlers] signup scene", "", reg, telelog.Trace(c))
	return sendSignupScene(c, reg, singles)
}

// coupleAdd handles the couple signup action
func (h *Handlers) coupleAdd(c tele.Context, eventID string, role models.Role, other any) error {
	u := h.userGet(c)

	reg, err := h.events.CoupleAdd(h.ctx(c), eventID, &u.Profile, role, other)
	if err != nil {
		h.log.Error("[handlers] failed to add couple: "+err.Error(),
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] couple add", "", reg, telelog.Trace(c))

	// if the result is retryable, update the session
	var singles []models.SessionSingle
	if reg.Result.IsRetryable() {
		singles = fmtSingles(reg.Event.Singles, role.Opposite())
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
	return sendResult(c, reg, singles)
}

// singleAdd handles the single signup action
func (h *Handlers) singleAdd(c tele.Context, eventID string, role models.Role) error {
	u := h.userGet(c)
	profile := models.NewProfile(*c.Sender())

	reg, err := h.events.SingleAdd(h.ctx(c), eventID, &profile, role)
	if err != nil {
		h.log.Error("[handlers] failed to add single: "+err.Error(),
			"event_id", eventID,
			"profile", profile.LogValue(),
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] single add", "", reg, telelog.Trace(c))

	// if the result is retryable, update the session
	var singles []models.SessionSingle
	if reg.Result.IsRetryable() {
		singles = fmtSingles(reg.Event.Singles, role.Opposite())
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
	return sendResult(c, reg, singles)
}

// dancerRemove handles the dancer remove action
func (h *Handlers) dancerRemove(c tele.Context, eventID string) error {
	u := h.userGet(c)
	profile := models.NewProfile(*c.Sender())

	reg, err := h.events.DancerRemove(h.ctx(c), eventID, &profile)
	if err != nil {
		h.log.Error("[handlers] failed to remove dancer: "+err.Error(),
			"event_id", eventID,
			"profile", profile.LogValue(),
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] dancer remove", "", reg, telelog.Trace(c))

	u.Session = models.Session{}
	h.userUpsert(c, u)
	return sendResult(c, reg, nil)
}

// sendErr sends an error message.
// It resets user session and removes the reply keyboard.
func (h *Handlers) sendErr(c tele.Context, msg string) error {
	// clear user session
	u := h.userGet(c)
	u.Session = models.Session{}
	h.userUpsert(c, u)
	return c.Send(msg, tele.RemoveKeyboard)
}
