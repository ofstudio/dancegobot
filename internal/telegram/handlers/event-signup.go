package handlers

import (
	"regexp"
	"strconv"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// SignupUserShared - handles the user shared event.
func (h *Handlers) SignupUserShared(c tele.Context) error {
	h.log.Info("[handlers] shared user received", telelog.Attr(c))
	u := h.userGet(c)
	if u.Session.Action != models.SessionSignup {
		h.log.Error("[handlers] unexpected shared user", telelog.Trace(c))
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

// SignupPartnerLegacy - handles the /partner command (legacy).
// This is to provide familiar user experience with github.com/Tayrinn/CoopDance bot.
// The payload will be treated as a text message.
func (h *Handlers) SignupPartnerLegacy(c tele.Context) error {
	h.log.Info("[handlers] /partner received: payload will be treated as text message", telelog.Attr(c))
	if c.Message().Payload == "" {
		return nil
	}
	c.Message().Text = c.Message().Payload
	c.Message().Payload = ""
	return h.Text(c)
}

// signupText handles the text message for the signup scene.
func (h *Handlers) signupText(c tele.Context) error {
	u := h.userGet(c)
	text := c.Text()
	switch {
	case text == locale.BtnDancerRemove:
		return h.dancerRemove(c, u.Session.EventID)
	case text == locale.BtnAsSingle[u.Session.Role]:
		return h.singleAdd(c, u.Session.EventID, u.Session.Role)
	case h.isLikeSingleCaption(text):
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

func (h *Handlers) signupScene(c tele.Context, eventID string, role models.Role) error {
	u := h.userGet(c)
	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] signup scene: failed to get event: "+err.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	reg := h.eventService.RegistrationGet(event, &u.Profile, role)
	if reg == nil {
		h.log.Error("[handlers] signup scene: failed to get registration",
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			"role", role.String(),
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}

	// if the dancer can register or already registered or the event is not closed for all
	// update the session with the singles
	var singles []models.SessionSingle
	if reg.Status.CanRegister() || reg.Status.IsRegistered() || !event.Settings.Closed {
		singles = h.fmtSingles(event.Singles, role.Opposite())
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
	h.userUpdateSession(c, u)

	if event.Settings.Closed {
		return c.Send(locale.ResultEventClosed, tele.RemoveKeyboard, tele.NoPreview, tele.ModeHTML)
	}

	if event.Removed {
		return c.Send(locale.ResultEventRemoved, tele.RemoveKeyboard, tele.NoPreview, tele.ModeHTML)
	}

	h.log.Info("[handlers] signup scene", "", reg, telelog.Trace(c))
	return views.SendSignupScene(c, reg, singles)
}

// coupleAdd handles the couple signup action
func (h *Handlers) coupleAdd(c tele.Context, eventID string, role models.Role, other any) error {
	u := h.userGet(c)

	reg, err := h.eventService.CoupleAdd(h.ctx(c), eventID, &u.Profile, role, other)
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
		singles = h.fmtSingles(reg.Event.Singles, role.Opposite())
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
	h.userUpdateSession(c, u)
	return views.SendResult(c, reg, singles)
}

// singleAdd handles the single signup action
func (h *Handlers) singleAdd(c tele.Context, eventID string, role models.Role) error {
	u := h.userGet(c)
	profile := models.NewProfile(*c.Sender())

	reg, err := h.eventService.SingleAdd(h.ctx(c), eventID, &profile, role)
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
		singles = h.fmtSingles(reg.Event.Singles, role.Opposite())
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
	h.userUpdateSession(c, u)
	return views.SendResult(c, reg, singles)
}

// dancerRemove handles the dancer remove action
func (h *Handlers) dancerRemove(c tele.Context, eventID string) error {
	u := h.userGet(c)
	reg, err := h.eventService.DancerRemove(h.ctx(c), eventID, &u.Profile)
	if err != nil {
		h.log.Error("[handlers] failed to remove dancer: "+err.Error(),
			"event_id", eventID,
			"profile", u.Profile.LogValue(),
			telelog.Trace(c))
		return h.sendErr(c, locale.ErrSomethingWrong)
	}
	h.log.Info("[handlers] dancer remove", "", reg, telelog.Trace(c))

	u.Session = models.Session{}
	h.userUpdateSession(c, u)
	return views.SendResult(c, reg, nil)
}

// fmtSingles formats the singles for the signup scene.
// Returns the list of profiles with reply button captions.
// Caption format: "1. Full Name (@username)"
// or just "1. Full Name" if no Telegram username.
func (h *Handlers) fmtSingles(singles []models.Dancer, role models.Role) []models.SessionSingle {
	var s []models.SessionSingle
	for i, d := range singles {
		if d.Profile == nil {
			continue
		}
		if d.Role == role {
			caption := strconv.Itoa(i+1) + ". " + d.FullName
			if d.Profile.Username != "" {
				caption += " (@" + d.Profile.Username + ")"
			}
			s = append(s, models.SessionSingle{
				Caption: caption,
				Profile: *d.Profile,
			})

		}
	}
	return s
}

var reSingleCapt = regexp.MustCompile(`^\d+\. .+$`)

// isLikeSingleCaption checks if the text looks like a single button caption.
func (h *Handlers) isLikeSingleCaption(text string) bool {
	return reSingleCapt.MatchString(text)
}
