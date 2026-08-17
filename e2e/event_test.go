package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/app"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestEventDraft(t *testing.T) {
	t.Run("empty inline query", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		require.Equal(t, locale.QueryTextEmpty, req.JSON.Get("results.0.title").String())
		require.Equal(t, locale.QueryDescriptionEmpty, req.JSON.Get("results.0.description").String())
		require.Equal(t, locale.QueryTextEmpty, req.JSON.Get("results.0.message_text").String())
	})

	t.Run("whitespace inline query", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "   ", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		require.Equal(t, locale.QueryTextEmpty, req.JSON.Get("results.0.title").String())
		require.Equal(t, locale.QueryDescriptionEmpty, req.JSON.Get("results.0.description").String())
	})

	t.Run("non-empty inline query", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Test text", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		eventID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Test text", req.JSON.Get("results.0.title").String())
		require.Equal(t, locale.QueryDescription, req.JSON.Get("results.0.description").String())
		require.Equal(t, "Test text", req.JSON.Get("results.0.input_message_content.message_text").String())
		require.Equal(t, "HTML", req.JSON.Get("results.0.input_message_content.parse_mode").String())
		require.Contains(t, req.JSON.Get("results.0.reply_markup.inline_keyboard").Raw, "signup|"+eventID+"|leader")

		event := env.eventGet(eventID)
		require.Equal(t, "Test text", event.Caption)
	})

	t.Run("html special chars are plain text", func(t *testing.T) {
		env := newEnv(t)
		text := "Танцы <tag> & friends"
		escaped := "Танцы &lt;tag&gt; &amp; friends"
		inlineMessageID := "test-inline-message-html-caption"

		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: text, ChatType: "supergroup"}))
		req := env.tg.Wait("answerInlineQuery")
		eventID := req.JSON.Get("results.0.id").String()
		require.Equal(t, text, req.JSON.Get("results.0.title").String())
		require.Equal(t, escaped, req.JSON.Get("results.0.input_message_content.message_text").String())

		env.process(env.inlineResult(tele.InlineResult{
			Sender:    userJohn,
			ResultID:  eventID,
			Query:     text,
			MessageID: inlineMessageID,
		}))
		edit := env.tg.Wait("editMessageText")
		require.Equal(t, inlineMessageID, edit.String("inline_message_id"))
		require.Equal(t, escaped+"\n\n", edit.String("text"))
		require.Equal(t, text, env.eventGet(eventID).Caption)
	})

	t.Run("inline query with couple limit", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Limited event /2", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		eventID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Limited event", req.JSON.Get("results.0.title").String())
		require.Equal(t, "Лимит 2 пары", req.JSON.Get("results.0.description").String())
		require.Equal(t, 2, env.eventGet(eventID).Settings.Limit)
	})

	t.Run("date does not set limit", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Class 3/4/2026", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		eventID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Class 3/4/2026", req.JSON.Get("results.0.title").String())
		require.Equal(t, 0, env.eventGet(eventID).Settings.Limit)
	})

	t.Run("unsupported limit shortcuts are ordinary text", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Limited event /0 /100", ChatType: "supergroup"}))

		req := env.tg.Wait("answerInlineQuery")
		eventID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Limited event /0 /100", req.JSON.Get("results.0.title").String())
		require.Equal(t, 0, env.eventGet(eventID).Settings.Limit)
	})

	t.Run("long inline query warnings", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: strings.Repeat("A", 250), ChatType: "supergroup"}))
		require.Equal(t, "Осталось 5 символов", env.tg.Wait("answerInlineQuery").JSON.Get("results.0.description").String())

		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: strings.Repeat("A", 300), ChatType: "supergroup"}))
		require.Equal(t, locale.QueryOverflow, env.tg.Wait("answerInlineQuery").JSON.Get("results.0.description").String())
	})
}

func TestEventPostAdd(t *testing.T) {
	t.Run("via chosen inline result", func(t *testing.T) {
		env := newEnv(t)
		eventID := env.eventDraftCreate(queryA)

		env.process(env.inlineResult(tele.InlineResult{
			Sender:    userJohn,
			ResultID:  eventID,
			Query:     queryA.Text,
			MessageID: "test-inline-message-ChosenInlineResult",
		}))

		edit := env.tg.Wait("editMessageText")
		require.Equal(t, "test-inline-message-ChosenInlineResult", edit.String("inline_message_id"))
		require.Equal(t, "test-inline-message-ChosenInlineResult", env.eventGet(eventID).Post.InlineMessageID)
	})

	t.Run("via callback query success", func(t *testing.T) {
		env := newEnv(t)
		eventID := env.eventDraftCreate(queryB)

		env.process(env.callback(tele.Callback{
			Sender:    userJohn,
			MessageID: "test-inline-message-CallbackQuery",
			Data:      "\fsignup|" + eventID + "|leader|rand-token",
		}))

		resp := env.tg.Wait("answerCallbackQuery")
		require.Regexp(t, rxUrlSignupLeader, resp.String("url"))
		env.tg.Wait("editMessageText")
		require.Equal(t, "test-inline-message-CallbackQuery", env.eventGet(eventID).Post.InlineMessageID)
	})

	t.Run("via callback query invalid", func(t *testing.T) {
		env := newEnv(t)
		eventID := env.eventDraftCreate(queryB)

		env.process(env.callback(tele.Callback{
			Sender:    userJohn,
			MessageID: "test-inline-message-CallbackQuery",
			Data:      "\fsignup|INVALID_PAYLOAD",
		}))

		resp := env.tg.Wait("answerCallbackQuery")
		require.Equal(t, locale.ErrSomethingWrong, resp.String("text"))
		require.Nil(t, env.eventGet(eventID).Post)
	})
}

func TestPostChatAdd(t *testing.T) {
	t.Run("if before chosen inline result", func(t *testing.T) {
		env := newEnv(t)
		eventID := env.eventDraftCreate(queryA)

		env.process(env.channelPost(tele.Message{
			ID:     12345,
			Sender: userJohn,
			Chat:   chatSuperGroup,
			Text:   queryA.Text,
			Via:    botUser,
			ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
				{Text: locale.RoleIcon[models.RoleLeader], Data: "\fsignup|" + eventID + "|leader|rand-token"},
			}}},
		}))

		env.process(env.inlineResult(tele.InlineResult{
			Sender:    userJohn,
			ResultID:  eventID,
			Query:     queryA.Text,
			MessageID: "test-inline-message",
		}))
		env.tg.Wait("editMessageText")

		event := env.eventGet(eventID)
		require.Equal(t, "test-inline-message", event.Post.InlineMessageID)
		require.Equal(t, int64(-1001234567890), event.Post.Chat.ID)
		require.Equal(t, 12345, event.Post.ChatMessageID)
	})

	t.Run("if after chosen inline result", func(t *testing.T) {
		env := newEnv(t)
		eventID := env.eventDraftCreate(queryB)

		env.process(env.inlineResult(tele.InlineResult{
			Sender:    userJohn,
			ResultID:  eventID,
			Query:     queryB.Text,
			MessageID: "test-inline-message",
		}))
		firstEdit := env.tg.Wait("editMessageText")

		env.process(env.channelPost(tele.Message{
			ID:     67890,
			Sender: userJane,
			Chat:   chatSuperGroup,
			Text:   queryB.Text,
			Via:    botUser,
			ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
				{Text: locale.RoleIcon[models.RoleLeader], URL: "https://t.me/" + botUser.Username + "?start=rand-signup-" + eventID + "-leader"},
			}}},
		}))
		secondEdit := env.tg.Wait("editMessageText")
		require.Equal(t, firstEdit.InlineKeyboardRaw(), secondEdit.InlineKeyboardRaw())
		require.Contains(t, secondEdit.InlineKeyboardRaw(), "?start=v1-signup-"+eventID+"-leader")

		event := env.eventGet(eventID)
		require.Equal(t, int64(-1001234567890), event.Post.Chat.ID)
		require.Equal(t, 67890, event.Post.ChatMessageID)
	})
}

func TestEventRemovedAfterRenderFailures(t *testing.T) {
	env := newEnv(t)
	eventID := env.eventDraftCreate(queryA)

	for i := 0; i < 3; i++ {
		env.tg.RespondError("editMessageText", 400, "Bad Request: MESSAGE_ID_INVALID")
		env.process(env.inlineResult(tele.InlineResult{
			Sender:    userJohn,
			ResultID:  eventID,
			Query:     queryA.Text,
			MessageID: "missing-inline-message",
		}))
		env.tg.Wait("editMessageText")
	}

	require.Eventually(t, func() bool {
		event := env.eventGet(eventID)
		return event.RenderFails == 3 && event.Removed
	}, time.Second, 10*time.Millisecond)
}

func TestLongPollerSmoke(t *testing.T) {
	tg := teletest.New(t, teletest.WithBotUser(botUser))
	tg.Ignore("getMe", "setMyCommands", "deleteWebhook", "getUpdates")

	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.Default()
	cfg.Bot.ApiURL = tg.URL()
	cfg.Bot.Token = "123456:ABCDEF"
	cfg.Bot.RPS = 1000
	cfg.Bot.Timeout = 50 * time.Millisecond
	cfg.DB.Filepath = t.TempDir() + "/app_test.db"
	cfg.RendererRepeats = []time.Duration{}
	cfg.ReRenderOnStartup = 0
	cfg.DraftCleanupEvery = 0
	cfg.DraftCleanupOlderThan = 0

	a := app.New(cfg)
	done := make(chan error, 1)
	go func() { done <- a.Start(ctx) }()
	t.Cleanup(func() {
		cancel()
		require.NoError(t, <-done)
		tg.AssertNoUnexpected()
	})

	tg.Wait("getUpdates", 3*time.Second)
	tg.Push(tg.Message(&tele.Message{
		Sender: userJohn,
		Chat:   privateChat(userJohn),
		Text:   "/start",
	}))
	req := tg.Wait("sendMessage", 3*time.Second)
	require.Equal(t, userJohn.ID, req.ChatIDInt())
	require.Contains(t, req.String("text"), "Привет! Это бот")
	require.Contains(t, req.String("text"), "@test_bot")
}
