package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestNewEventSubscriptionNotificationLifecycle(t *testing.T) {
	env := newEnv(t)
	first := testEvent("subscription_source", "First event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), first))
	publishSubscriptionEvent(env, first)

	env.process(env.message(userJohn, "/start v1-subscribe-"+first.ID))
	prompt := env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	require.Contains(t, prompt.InlineKeyboardRaw(), locale.BtnSubscribe)
	require.Contains(t, prompt.InlineKeyboardRaw(), locale.BtnClose)
	require.Contains(t, prompt.InlineKeyboardRaw(), first.ID)

	status, err := env.app.Services.Subscription.Status(
		context.Background(),
		env.eventGet(first.ID),
		models.NewProfile(*userJohn),
	)
	require.NoError(t, err)
	require.Equal(t, services.SubscriptionStatus{Available: true}, status)

	confirm := tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 776, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + first.ID,
	}
	env.process(env.callback(confirm))
	edit := env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	require.Empty(t, edit.InlineKeyboardRaw())
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	// A stale repeated callback remains idempotent and uses the same success text.
	env.tg.RespondError("editMessageText", tele.ErrMessageNotModified.Code, tele.ErrMessageNotModified.Description)
	env.process(env.callback(confirm))
	edit = env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	require.Empty(t, edit.InlineKeyboardRaw())
	answer = env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))

	second := testEvent("subscription_target", "Second event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), second))
	publishSubscriptionEvent(env, second)
	notification := env.waitSendMessage(userJohn.ID, "🔔 Новая запись в Test Super Group.")
	require.Contains(t, notification.InlineKeyboardRaw(), locale.BtnChatLink)
	require.Contains(t, notification.InlineKeyboardRaw(), locale.BtnUnsubscribe)
	require.Contains(t, notification.InlineKeyboardRaw(), second.ID)

	// A repeated Telegram update may invoke PostChatAdd again, but the business
	// notification is claimed once through Event.SubscribersNotified.
	publishSubscriptionEvent(env, second)

	unsubscribe := tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 777, Chat: privateChat(userJohn)},
		Data:    "\fnotification_unsubscribe|" + second.ID,
	}
	env.process(env.callback(unsubscribe))
	edit = env.waitEditMessageText("Вы отписались от Test Super Group.")
	require.Contains(t, edit.String("text"), locale.LinkResubscribe)
	require.Contains(t, edit.String("text"), "v1-subscribe-"+second.ID)
	require.Empty(t, edit.InlineKeyboardRaw())
	answer = env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	env.tg.RespondError("editMessageText", tele.ErrSameMessageContent.Code, tele.ErrSameMessageContent.Description)
	env.process(env.callback(unsubscribe))
	edit = env.waitEditMessageText("Вы отписались от Test Super Group.")
	require.Empty(t, edit.InlineKeyboardRaw())
	answer = env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	env.process(env.message(userJohn, "/start v1-subscribe-"+second.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 779, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + second.ID,
	}))
	env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	env.tg.Wait("answerCallbackQuery")

	status, err = env.app.Services.Subscription.Status(context.Background(), env.eventGet(second.ID), models.NewProfile(*userJohn))
	require.NoError(t, err)
	require.Equal(t, services.SubscriptionStatus{Available: true, Subscribed: true}, status)
}

func TestNewEventSubscriptionNotificationLifecycleWithoutBotAdmin(t *testing.T) {
	env := newEnv(t, teletest.WithBotAdministrator(false))
	first := testEvent("subscription_non_admin_source", "First event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), first))
	publishSubscriptionEvent(env, first)

	env.process(env.message(userJohn, "/start v1-subscribe-"+first.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 790, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + first.ID,
	}))
	env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	env.tg.Wait("answerCallbackQuery")

	second := testEvent("subscription_non_admin_target", "Second event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), second))
	publishSubscriptionEvent(env, second)
	notification := env.waitSendMessage(userJohn.ID, "🔔 Новая запись в Test Super Group.")
	require.Contains(t, notification.InlineKeyboardRaw(), locale.BtnChatLink)
	require.Contains(t, notification.InlineKeyboardRaw(), locale.BtnUnsubscribe)

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 791, Chat: privateChat(userJohn)},
		Data:    "\fnotification_unsubscribe|" + second.ID,
	}))
	edit := env.waitEditMessageText("Вы отписались от Test Super Group.")
	require.Contains(t, edit.String("text"), locale.LinkResubscribe)
	require.Empty(t, edit.InlineKeyboardRaw())
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
}

func TestSubscriptionControlsAfterSignupAndInMy(t *testing.T) {
	env := newEnv(t)
	event := testEvent("subscription_signup", "Subscription signup", userJane)
	event.Settings.AutoPairing = true
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
	publishSubscriptionEvent(env, event)

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
	env.process(env.message(userJohn, locale.BtnAsSingle[models.RoleLeader]))
	env.tg.Wait("editMessageText")
	result := env.waitSendMessage(userJohn.ID, "Добавил тебя в список ищущих пару")
	require.Contains(t, result.String("text"), `>`+locale.LinkSubscribe+`</a>`)
	require.Contains(t, result.String("text"), "v1-subscribe-"+event.ID)

	env.process(env.message(userJohn, "/start v1-subscribe-"+event.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 777, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_close",
	}))
	deleted := env.tg.Wait("deleteMessage")
	require.Equal(t, 777, deleted.Int("message_id"))
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))

	env.process(env.message(userJohn, "/my"))
	my := env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID && strings.Contains(req.InlineKeyboardRaw(), locale.BtnSubscribe)
	})
	require.Contains(t, my.InlineKeyboardRaw(), event.ID)

	subscribe := tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 778, Chat: privateChat(userJohn)},
		Data:    "\fmy_subscription|subscribe|" + event.ID + "|0",
	}
	env.process(env.callback(subscribe))
	updatedMy := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return strings.Contains(req.InlineKeyboardRaw(), locale.BtnUnsubscribe)
	})
	require.Contains(t, updatedMy.InlineKeyboardRaw(), event.ID)
	answer = env.tg.Wait("answerCallbackQuery")
	require.Equal(t, "Вы подписались на новые мероприятия в Test Super Group.", answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	env.tg.RespondError("editMessageText", tele.ErrSameMessageContent.Code, tele.ErrSameMessageContent.Description)
	env.process(env.callback(subscribe))
	updatedMy = env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return strings.Contains(req.InlineKeyboardRaw(), locale.BtnUnsubscribe)
	})
	require.Contains(t, updatedMy.InlineKeyboardRaw(), event.ID)
	answer = env.tg.Wait("answerCallbackQuery")
	require.Equal(t, "Вы подписались на новые мероприятия в Test Super Group.", answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 778, Chat: privateChat(userJohn)},
		Data:    "\fmy_subscription|unsubscribe|" + event.ID + "|0",
	}))
	updatedMy = env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return strings.Contains(req.InlineKeyboardRaw(), locale.BtnSubscribe)
	})
	require.Contains(t, updatedMy.InlineKeyboardRaw(), event.ID)
	answer = env.tg.Wait("answerCallbackQuery")
	require.Equal(t, "Вы отписались от Test Super Group.", answer.String("text"))
	require.False(t, answer.Bool("show_alert"))

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 778, Chat: privateChat(userJohn)},
		Data:    "\fmy_subscription|invalid|" + event.ID + "|0",
	}))
	answer = env.tg.Wait("answerCallbackQuery")
	require.Equal(t, locale.ErrSomethingWrong, answer.String("text"))
	require.False(t, answer.Bool("show_alert"))
}

func TestSubscriptionConfirmationUnavailable(t *testing.T) {
	env := newEnv(t)
	event := testEvent("subscription_unavailable", "Unavailable subscription", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
	publishSubscriptionEvent(env, event)
	env.tg.WaitFor("getChatMember", func(req teletest.Request) bool {
		return req.JSON.Get("user_id").Int() == botUser.ID
	})

	env.process(env.message(userJohn, "/start v1-subscribe-"+event.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	env.tg.Respond("getChatMember", &tele.ChatMember{
		User: botUser,
		Role: tele.Administrator,
	})
	env.tg.RespondError("getChatMember", 400, "Bad Request: member list is inaccessible")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 780, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + event.ID,
	}))
	edit := env.waitEditMessageText(locale.SubscriptionUnavailable)
	require.Empty(t, edit.InlineKeyboardRaw())
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
}

func TestSubscriptionBotMembershipErrorUsesRelaxedPolicy(t *testing.T) {
	env := newEnv(t)
	event := testEvent("subscription_relaxed_error", "Relaxed subscription", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
	publishSubscriptionEvent(env, event)
	env.tg.WaitFor("getChatMember", func(req teletest.Request) bool {
		return req.JSON.Get("user_id").Int() == botUser.ID
	})

	env.tg.RespondError("getChatMember", 400, "Bad Request: member list is inaccessible")
	env.process(env.message(userJohn, "/start v1-subscribe-"+event.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")

	env.tg.RespondError("getChatMember", 400, "Bad Request: member list is inaccessible")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 792, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + event.ID,
	}))
	env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
}

func TestNotificationUnsubscribeWithoutResubscribe(t *testing.T) {
	env := newEnv(t)
	first := testEvent("subscription_no_resub_source", "First event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), first))
	publishSubscriptionEvent(env, first)
	subscribeToEvent(env, first, 781)

	second := testEvent("subscription_no_resub_target", "Second event", userJane)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), second))
	publishSubscriptionEvent(env, second)
	env.waitSendMessage(userJohn.ID, "🔔 Новая запись в Test Super Group.")

	env.tg.Respond("getChatMember", &tele.ChatMember{
		User: botUser,
		Role: tele.Administrator,
	})
	env.tg.RespondError("getChatMember", 400, "Bad Request: member list is inaccessible")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 782, Chat: privateChat(userJohn)},
		Data:    "\fnotification_unsubscribe|" + second.ID,
	}))
	edit := env.waitEditMessageText("Вы отписались от Test Super Group.")
	require.NotContains(t, edit.String("text"), locale.LinkResubscribe)
	require.Empty(t, edit.InlineKeyboardRaw())
	answer := env.tg.Wait("answerCallbackQuery")
	require.Empty(t, answer.String("text"))
}

func subscribeToEvent(env *testEnv, event *models.Event, messageID int) {
	env.t.Helper()
	env.process(env.message(userJohn, "/start v1-subscribe-"+event.ID))
	env.waitSendMessage(userJohn.ID, "Подписаться на Test Super Group?")
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: messageID, Chat: privateChat(userJohn)},
		Data:    "\fsubscription_subscribe|" + event.ID,
	}))
	env.waitEditMessageText("Вы подписались на новые мероприятия в Test Super Group.")
	env.tg.Wait("answerCallbackQuery")
}

func publishSubscriptionEvent(env *testEnv, event *models.Event) {
	env.t.Helper()
	env.process(env.channelPost(tele.Message{
		ID:     500,
		Sender: userJane,
		Chat:   chatSuperGroup,
		Text:   event.Caption,
		Via:    botUser,
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{{
			Text: locale.RoleIcon[models.RoleLeader],
			Data: "\fsignup|" + event.ID + "|leader|v1",
		}}}},
	}))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID
	})
}
