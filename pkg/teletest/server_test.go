package teletest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"
)

func TestServerRecordsRequests(t *testing.T) {
	tg := New(t)
	tg.Ignore("getMe")

	bot, err := tele.NewBot(tele.Settings{
		URL:         tg.URL(),
		Token:       "123:ABC",
		Synchronous: true,
	})
	require.NoError(t, err)

	_, err = bot.Send(&tele.User{ID: 42}, "Hello")
	require.NoError(t, err)

	req := tg.Wait("sendMessage")
	require.Equal(t, int64(42), req.ChatIDInt())
	require.Equal(t, "Hello", req.String("text"))
	tg.AssertNoUnexpected()
}

func TestServerPushesUpdatesForLongPoller(t *testing.T) {
	tg := New(t)
	tg.Push(tg.Message(&tele.Message{
		Sender: &tele.User{ID: 42, FirstName: "Alice"},
		Chat:   &tele.Chat{ID: 42, Type: tele.ChatPrivate},
		Text:   "/start",
	}))

	bot, err := tele.NewBot(tele.Settings{
		URL:         tg.URL(),
		Token:       "123:ABC",
		Poller:      &tele.LongPoller{Timeout: 100 * time.Millisecond},
		Synchronous: true,
	})
	require.NoError(t, err)

	done := make(chan struct{})
	bot.Handle("/start", func(c tele.Context) error {
		close(done)
		return nil
	})

	go bot.Start()
	defer bot.Stop()

	require.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestServerQueuedErrorResponse(t *testing.T) {
	tg := New(t)
	tg.Ignore("getMe")
	tg.RespondError("sendMessage", 400, "Bad Request: chat not found")

	bot, err := tele.NewBot(tele.Settings{
		URL:         tg.URL(),
		Token:       "123:ABC",
		Synchronous: true,
	})
	require.NoError(t, err)

	_, err = bot.Send(&tele.User{ID: 42}, "Hello")
	require.Error(t, err)
	require.Contains(t, err.Error(), "chat not found")

	tg.Wait("sendMessage")
	tg.AssertNoUnexpected()
}

func TestServerBotAdministratorOption(t *testing.T) {
	tg := New(t, WithBotAdministrator(false))
	tg.Ignore("getMe", "getChatMember")

	bot, err := tele.NewBot(tele.Settings{
		URL:         tg.URL(),
		Token:       "123:ABC",
		Synchronous: true,
	})
	require.NoError(t, err)

	member, err := bot.ChatMemberOf(&tele.Chat{ID: -1001}, bot.Me)
	require.NoError(t, err)
	require.Equal(t, tele.Member, member.Role)
	tg.AssertNoUnexpected()
}
