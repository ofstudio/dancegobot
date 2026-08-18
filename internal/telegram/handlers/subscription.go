package handlers

import (
	"errors"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

func (h *Handlers) subscribeFromLink(c tele.Context, eventID string) error {
	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		h.log.Error("[handlers] subscribe link: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.SubscriptionUnavailable)
	}
	if event == nil {
		return h.sendErr(c, locale.SubscriptionUnavailable)
	}
	status, err := h.subscriptionService.Status(h.ctx(c), event, h.userGet(c).Profile)
	if err != nil {
		h.log.Error("[handlers] subscribe link: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return h.sendErr(c, locale.SubscriptionUnavailable)
	}
	if !status.Available {
		return h.sendErr(c, locale.SubscriptionUnavailable)
	}
	return views.SubscriptionPrompt(c, event.Post.Chat, eventID)
}

func (h *Handlers) subscriptionURL(c tele.Context, reg models.Registration) string {
	if reg.Result != models.ResultRegisteredAsSingle && reg.Result != models.ResultRegisteredInCouple {
		return ""
	}
	event := reg.Event
	if event == nil {
		return ""
	}
	status, err := h.subscriptionService.Status(h.ctx(c), event, h.userGet(c).Profile)
	if err != nil {
		h.log.Error("[handlers] failed to get subscription status: "+err.Error(),
			"event_id", event.ID,
			telelog.Trace(c))
		return ""
	}
	if !status.Available || status.Subscribed {
		return ""
	}
	return views.EventSubscribeURL(event.ID)
}

// CbSubscriptionSubscribe handles subscription confirmation from a deep link.
func (h *Handlers) CbSubscriptionSubscribe(c tele.Context) error {
	h.log.Info("[handlers] subscription confirmation callback received", telelog.Attr(c))
	if len(c.Args()) < 1 {
		return c.RespondText(locale.ErrSomethingWrong)
	}
	eventID := c.Args()[0]
	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		return h.subscriptionConfirmationUnavailable(c, err, eventID)
	}
	if event == nil || event.Post == nil || event.Post.Chat == nil {
		return h.subscriptionConfirmationUnavailable(c, nil, eventID)
	}
	_, _, err = h.subscriptionService.Subscribe(h.ctx(c), eventID, h.userGet(c).Profile)
	if err != nil {
		return h.subscriptionConfirmationUnavailable(c, err, eventID)
	}
	if err = views.SubscriptionSubscribed(c, event.Post.Chat); err != nil {
		h.log.Error("[handlers] subscription confirmation: "+err.Error(), "event_id", eventID, telelog.Trace(c))
		return c.RespondText(locale.ErrSomethingWrong)
	}
	return c.Respond()
}

// CbSubscriptionClose handles closing a deep-link subscription confirmation.
func (h *Handlers) CbSubscriptionClose(c tele.Context) error {
	h.log.Info("[handlers] subscription confirmation close callback received", telelog.Attr(c))
	if err := c.Delete(); err != nil {
		h.log.Error("[handlers] subscription confirmation close: "+err.Error(), telelog.Trace(c))
		return c.RespondText(locale.ErrSomethingWrong)
	}
	return c.Respond()
}

// CbNotificationUnsubscribe handles unsubscription from a new-event notification.
func (h *Handlers) CbNotificationUnsubscribe(c tele.Context) error {
	h.log.Info("[handlers] notification unsubscribe callback received", telelog.Attr(c))
	if len(c.Args()) < 1 {
		return c.RespondText(locale.ErrSomethingWrong)
	}
	eventID := c.Args()[0]
	event, err := h.eventService.Get(h.ctx(c), eventID)
	if err != nil {
		return c.RespondText(locale.SubscriptionUnavailable)
	}
	if event == nil || event.Post == nil || event.Post.Chat == nil {
		return c.RespondText(locale.SubscriptionUnavailable)
	}
	profile := h.userGet(c).Profile
	_, _, err = h.subscriptionService.Unsubscribe(h.ctx(c), eventID, profile)
	if err != nil {
		return h.respondSubscriptionError(c, err, "notification unsubscribe")
	}
	subscribeURL := h.subscriptionURLIfAvailable(c, event, profile)
	if err = views.SubscriptionUnsubscribed(c, event.Post.Chat, subscribeURL); err != nil {
		h.log.Error("[handlers] notification unsubscribe: "+err.Error(), telelog.Trace(c))
		return c.RespondText(locale.ErrSomethingWrong)
	}
	return c.Respond()
}

func (h *Handlers) subscriptionURLIfAvailable(
	c tele.Context,
	event *models.Event,
	profile models.Profile,
) string {
	status, err := h.subscriptionService.Status(h.ctx(c), event, profile)
	if err != nil {
		h.log.Error("[handlers] notification unsubscribe status: "+err.Error(),
			"event_id", event.ID,
			telelog.Trace(c))
		return ""
	}
	if !status.Available {
		return ""
	}
	return views.EventSubscribeURL(event.ID)
}

func (h *Handlers) subscriptionConfirmationUnavailable(c tele.Context, err error, eventID string) error {
	if err != nil {
		if !errors.Is(err, services.ErrSubscriptionUnavailable) {
			h.log.Error("[handlers] subscription confirmation: "+err.Error(),
				"event_id", eventID,
				telelog.Trace(c))
		}
	}
	if editErr := views.SubscriptionUnavailable(c); editErr != nil {
		h.log.Error("[handlers] subscription confirmation: "+editErr.Error(),
			"event_id", eventID,
			telelog.Trace(c))
		return c.RespondText(locale.ErrSomethingWrong)
	}
	return c.Respond()
}

func (h *Handlers) respondSubscriptionError(c tele.Context, err error, operation string) error {
	if errors.Is(err, services.ErrSubscriptionUnavailable) {
		return c.RespondText(locale.SubscriptionUnavailable)
	}
	h.log.Error("[handlers] "+operation+": "+err.Error(), telelog.Trace(c))
	return c.RespondText(locale.ErrSomethingWrong)
}
