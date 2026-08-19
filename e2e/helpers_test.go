package e2e

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/app"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

type testEnv struct {
	t      *testing.T
	tg     *teletest.Server
	app    *app.App
	ctx    context.Context
	cancel context.CancelFunc
}

func newEnv(t *testing.T, opts ...teletest.Option) *testEnv {
	t.Helper()

	serverOpts := []teletest.Option{teletest.WithBotUser(botUser)}
	serverOpts = append(serverOpts, opts...)
	tg := teletest.New(t, serverOpts...)
	tg.Ignore("getMe", "setMyCommands", "deleteWebhook", "getUpdates", "getChatMember")

	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.Default()
	cfg.Bot.ApiURL = tg.URL()
	cfg.Bot.Token = "123456:ABCDEF"
	cfg.Bot.Synchronous = true
	cfg.Bot.RPS = 1000
	cfg.Bot.Timeout = 5 * time.Second
	cfg.DB.Filepath = t.TempDir() + "/app_test.db"
	cfg.RendererRepeats = []time.Duration{}
	cfg.ReRenderOnStartup = 0
	cfg.DraftCleanupEvery = 0
	cfg.DraftCleanupOlderThan = 0

	a := app.New(cfg)
	require.NoError(t, a.Init(ctx))

	env := &testEnv{t: t, tg: tg, app: a, ctx: ctx, cancel: cancel}
	t.Cleanup(func() {
		cancel()
		a.Close()
		tg.AssertNoUnexpected()
	})
	return env
}

func (e *testEnv) process(update tele.Update) {
	e.t.Helper()
	e.app.Bot.ProcessUpdate(update)
}

func (e *testEnv) message(user *tele.User, text string) tele.Update {
	return e.tg.Message(&tele.Message{
		Sender: user,
		Chat:   privateChat(user),
		Text:   text,
	})
}

func (e *testEnv) inlineQuery(query tele.Query) tele.Update {
	q := query
	return e.tg.InlineQuery(&q)
}

func (e *testEnv) inlineResult(result tele.InlineResult) tele.Update {
	r := result
	return e.tg.InlineResult(&r)
}

func (e *testEnv) callback(callback tele.Callback) tele.Update {
	c := callback
	return e.tg.CallbackQuery(&c)
}

func (e *testEnv) channelPost(msg tele.Message) tele.Update {
	return tele.Update{ChannelPost: &msg}
}

func (e *testEnv) waitSendMessage(chatID int64, textContains string) teletest.Request {
	e.t.Helper()
	return e.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == chatID && strings.Contains(req.String("text"), textContains)
	})
}

func (e *testEnv) waitEditMessageText(textContains string) teletest.Request {
	e.t.Helper()
	return e.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return strings.Contains(req.String("text"), textContains)
	})
}

func (e *testEnv) eventDraftCreate(query tele.Query) string {
	e.t.Helper()
	e.process(e.inlineQuery(query))
	req := e.tg.Wait("answerInlineQuery")
	eventID := req.JSON.Get("results.0.id").String()
	require.Regexp(e.t, `^[a-zA-Z0-9]{12}$`, eventID)
	return eventID
}

func (e *testEnv) eventGet(id string) *models.Event {
	e.t.Helper()
	event, err := e.app.Services.Event.Get(context.Background(), id)
	require.NoError(e.t, err)
	require.NotNil(e.t, event)
	return event
}

func signupPayload(eventID string, role models.Role) string {
	return "rand-signup-" + eventID + "-" + role.String()
}

func privateChat(user *tele.User) *tele.Chat {
	return &tele.Chat{
		ID:        user.ID,
		Type:      tele.ChatPrivate,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}
}

func testEvent(id string, caption string, owner *tele.User) *models.Event {
	profile := models.NewProfile(*owner)
	return &models.Event{
		ID:        id,
		Caption:   caption,
		Post:      &models.Post{InlineMessageID: "inline_" + id},
		Owner:     profile,
		CreatedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
	}
}

func testLimitEvent(id string, owner *tele.User, limit int, couples ...models.Couple) *models.Event {
	chat := models.NewChat(chatSuperGroup)
	return &models.Event{
		ID:      id,
		Caption: "Limit event",
		Settings: models.EventSettings{
			Limit: limit,
		},
		Post: &models.Post{
			InlineMessageID: "inline_" + id,
			Chat:            &chat,
			ChatMessageID:   500,
		},
		Couples:   couples,
		Owner:     models.NewProfile(*owner),
		CreatedAt: limitTestBaseTime,
	}
}

func testLimitCouple(
	leader *tele.User,
	follower *tele.User,
	createdBy *tele.User,
	offset time.Duration,
	followerAsSingle bool,
) models.Couple {
	leaderProfile := models.NewProfile(*leader)
	followerProfile := models.NewProfile(*follower)
	return models.Couple{
		Dancers: []models.Dancer{
			{
				Profile:   &leaderProfile,
				FullName:  leaderProfile.FullName(),
				Role:      models.RoleLeader,
				CreatedAt: limitTestBaseTime.Add(offset),
			},
			{
				Profile:   &followerProfile,
				FullName:  followerProfile.FullName(),
				Role:      models.RoleFollower,
				AsSingle:  followerAsSingle,
				CreatedAt: limitTestBaseTime.Add(offset),
			},
		},
		CreatedBy: models.NewProfile(*createdBy),
		CreatedAt: limitTestBaseTime.Add(offset),
	}
}

var botUser = &tele.User{
	ID:        1234567890,
	IsBot:     true,
	FirstName: "test_bot",
	Username:  "test_bot",
}

var (
	userJohn = &tele.User{
		ID:        100,
		FirstName: "John",
	}
	userJane = &tele.User{
		ID:        101,
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "jane_doe",
	}
)

var chatSuperGroup = &tele.Chat{
	ID:    -1001234567890,
	Type:  tele.ChatSuperGroup,
	Title: "Test Super Group",
}

var (
	queryA = tele.Query{
		Sender:   userJohn,
		Text:     "Query A",
		ChatType: "supergroup",
	}
	queryB = tele.Query{
		Sender:   userJane,
		Text:     "Query B",
		ChatType: "channel",
	}
)

var (
	rxUrlSignupLeader = regexp.MustCompile(`^https://t.me/` + botUser.Username + `\?start=.+-signup-.+leader$`)
	limitTestBaseTime = time.Date(2024, 5, 6, 12, 0, 0, 0, time.UTC)
	limitUserAlice    = &tele.User{ID: 201, FirstName: "Alice"}
	limitUserBob      = &tele.User{ID: 202, FirstName: "Bob"}
	limitUserCarol    = &tele.User{ID: 203, FirstName: "Carol"}
	limitUserDan      = &tele.User{ID: 204, FirstName: "Dan"}
	limitUserEve      = &tele.User{ID: 205, FirstName: "Eve"}
)
