// Package teletest provides test helpers for Telegram bots built with
// gopkg.in/telebot.v4.
//
// The package starts a local fake Telegram Bot API server and records outgoing
// Bot API calls in a concurrency-safe request log. Tests can point telebot to
// Server.URL(), process updates directly with Bot.ProcessUpdate, and wait for
// the expected outgoing requests.
//
// A typical direct-update test looks like this:
//
//	tg := teletest.New(t, teletest.WithBotUser(&tele.User{
//		ID:       123,
//		IsBot:    true,
//		Username: "my_bot",
//	}))
//	defer tg.Close()
//
//	bot, err := tele.NewBot(tele.Settings{
//		URL:         tg.URL(),
//		Token:       "123:ABC",
//		Synchronous: true,
//	})
//	require.NoError(t, err)
//
//	bot.Handle("/start", func(c tele.Context) error {
//		return c.Send("Hello")
//	})
//
//	bot.ProcessUpdate(tg.Message(&tele.Message{
//		Sender: &tele.User{ID: 42, FirstName: "Alice"},
//		Chat:   &tele.Chat{ID: 42, Type: tele.ChatPrivate},
//		Text:   "/start",
//	}))
//
//	req := tg.Wait("sendMessage")
//	require.Equal(t, "42", req.String("chat_id"))
//	require.Equal(t, "Hello", req.String("text"))
//
// Long-polling code can be tested by pushing updates into the server:
//
//	tg.Push(tg.Message(&tele.Message{
//		Sender: &tele.User{ID: 42, FirstName: "Alice"},
//		Chat:   &tele.Chat{ID: 42, Type: tele.ChatPrivate},
//		Text:   "/start",
//	}))
//
// The fake server implements only the Bot API methods used by the tested bot.
// Unknown methods return {"ok":true,"result":true} by default and are still
// recorded, so tests can assert on them or report them as unexpected.
package teletest
