package app

import (
	"context"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/telegram/handlers"
	"github.com/ofstudio/dancegobot/internal/telegram/middleware"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
)

// wire initializes middleware and handlers and wires them to the bot.
func (a *App) wire(ctx context.Context) {

	// Initialize middleware and handlers
	m := middleware.
		NewMiddleware(a.cfg.Settings, a.EventService, a.UserService).
		WithLogger(a.log)
	h := handlers.
		NewHandlers(a.cfg.Settings, a.EventService, a.UserService, a.SubscriptionService).
		WithLogger(a.log)

	// Wire middleware to the bot
	a.Bot.Use(m.Context(ctx))
	a.Bot.Use(m.Trace())
	a.Bot.Use(m.Logger())
	a.Bot.Use(m.ChatMessage())
	a.Bot.Use(m.PassPrivateMessages())
	a.Bot.Use(m.User())

	// Bot commands
	a.Bot.Handle("/start", h.Start)
	a.Bot.Handle("/partner", h.SignupPartnerLegacy)
	a.Bot.Handle("/settings", h.UserSettingsScene)
	a.Bot.Handle("/my", h.My)

	// Handlers on generic events
	a.Bot.Handle(tele.OnText, h.Text)
	a.Bot.Handle(tele.OnUserShared, h.SignupUserShared)
	a.Bot.Handle(tele.OnQuery, h.EventQuery)
	a.Bot.Handle(tele.OnInlineResult, h.EventInlineResult)

	// Event signup buttons
	a.Bot.Handle(&views.BtnEventSignupCb, h.CbEventSignup)
	a.Bot.Handle(&views.BtnEventClosed, h.CbEventClosed)

	// User settings buttons
	a.Bot.Handle(&views.BtnUserSettingsAutoPair, h.CbUserSettingsAutoPair)
	a.Bot.Handle(&views.BtnUserSettingsLimit, h.CbUserSettingsLimitScene)
	a.Bot.Handle(&views.BtnUserSettingsLimitNum, h.CbUserSettingsLimitNum)
	a.Bot.Handle(&views.BtnUserSettingsHelp, h.CbUserSettingsHelp)
	a.Bot.Handle(&views.BtnUserSettingsBack, h.CbUserSettingsBack)

	// My scene buttons
	a.Bot.Handle(&views.BtnMyTurnPage, h.CbMyTurnPage)
	a.Bot.Handle(&views.BtnMyRefresh, h.CbMyRefresh)
	a.Bot.Handle(&views.BtnMySubscription, h.CbMySubscription)

	// Subscription buttons
	a.Bot.Handle(&views.BtnSubscriptionSubscribe, h.CbSubscriptionSubscribe)
	a.Bot.Handle(&views.BtnSubscriptionClose, h.CbSubscriptionClose)
	a.Bot.Handle(&views.BtnNotificationUnsubscribe, h.CbNotificationUnsubscribe)

	// Event settings buttons
	a.Bot.Handle(&views.BtnEventSettings, h.EventSettingsScene)
	a.Bot.Handle(&views.BtnEventSettingsBack, h.CbEventSettingsBack)
	a.Bot.Handle(&views.BtnEventSettingsAutoPair, h.CbEventSettingsToggles)
	a.Bot.Handle(&views.BtnEventSettingsClose, h.CbEventSettingsToggles)
	a.Bot.Handle(&views.BtnEventSettingsLimit, h.CbEventSettingsLimitScene)
	a.Bot.Handle(&views.BtnEventSettingsLimitNum, h.CbEventSettingsLimitNum)
	a.Bot.Handle(&views.BtnLimitChangedNotify, h.CbLimitChangedNotify)
	a.Bot.Handle(&views.BtnLimitChangedSkip, h.CbLimitChangedSkip)

	// This is needed to receive channel posts
	a.Bot.Handle(tele.OnChannelPost, func(_ tele.Context) error { return nil })
}
