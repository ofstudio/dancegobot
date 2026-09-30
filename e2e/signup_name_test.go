package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestSignupLongProfileNames(t *testing.T) {
	tests := []struct {
		scenario string
		limit    int
	}{
		{"single button", 64}, {"shared user", 64}, {"auto pairing", 64},
		{"single button", 8}, {"shared user", 8}, {"auto pairing", 8},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/limit=%d", tt.scenario, tt.limit), func(t *testing.T) {
			env := newEnvWithConfig(t, func(cfg *config.Config) { cfg.DancerNameMaxLen = tt.limit })
			leader := *userJohn
			leader.FirstName = strings.Repeat("Я", 40)
			leader.LastName = strings.Repeat("Ж", 40)
			follower := *userJane
			follower.FirstName = strings.Repeat("🙂", 40)
			follower.LastName = strings.Repeat("Ю", 40)
			leaderName := string([]rune(models.NewProfile(leader).FullName())[:tt.limit-1]) + locale.NameEllipsis
			followerName := string([]rune(models.NewProfile(follower).FullName())[:tt.limit-1]) + locale.NameEllipsis
			leaderLink := `<a href="tg://user?id=100">` + leaderName + `</a>`
			followerLink := `<a href="https://t.me/` + follower.Username + `">` + followerName + `</a>`
			event := testEvent("long_profile_signup", "Long profile signup", &leader)
			event.Settings.AutoPairing = tt.scenario == "auto pairing"
			require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

			env.process(env.message(&follower, "/start "+signupPayload(event.ID, models.RoleFollower)))
			env.waitSendMessage(follower.ID, locale.SignupNotRegistered)
			env.process(env.message(&follower, locale.BtnAsSingle[models.RoleFollower]))
			singlePost := env.tg.Wait("editMessageText")
			require.Contains(t, singlePost.String("text"), followerLink)
			require.NotContains(t, singlePost.String("text"), models.NewProfile(follower).FullName())
			env.waitSendMessage(follower.ID, "Добавил тебя в список ищущих пару")

			env.process(env.message(&leader, "/start "+signupPayload(event.ID, models.RoleLeader)))
			scene := env.waitSendMessage(leader.ID, locale.SignupNotRegistered)
			switch tt.scenario {
			case "single button":
				caption := "1. " + followerName + " (@" + follower.Username + ")"
				require.Equal(t, caption, scene.ReplyMarkup().Get("keyboard.1.0.text").String())
				env.process(env.message(&leader, caption))
			case "shared user":
				shared := &tele.RecipientShared{Users: []struct {
					UserID    int64       `json:"user_id"`
					FirstName string      `json:"first_name"`
					LastName  string      `json:"last_name"`
					Username  string      `json:"username"`
					Photo     *tele.Photo `json:"photo"`
				}{{
					UserID:    follower.ID,
					FirstName: follower.FirstName,
					LastName:  follower.LastName,
					Username:  follower.Username,
				}}}
				env.process(env.tg.Message(&tele.Message{Sender: &leader, Chat: privateChat(&leader), UserShared: shared}))
			case "auto pairing":
				env.process(env.message(&leader, locale.BtnAsSingle[models.RoleLeader]))
			}
			couplePost := env.tg.Wait("editMessageText")
			require.Contains(t, couplePost.String("text"), leaderLink)
			require.Contains(t, couplePost.String("text"), followerLink)
			require.NotContains(t, couplePost.String("text"), models.NewProfile(leader).FullName())
			require.NotContains(t, couplePost.String("text"), models.NewProfile(follower).FullName())
			result := env.waitSendMessage(leader.ID, "Вы зарегистрировались в паре")
			require.Contains(t, result.String("text"), followerLink)
			notification := env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
				return req.ChatIDInt() == follower.ID && strings.Contains(req.String("text"), leaderLink)
			})
			require.NotContains(t, notification.String("text"), models.NewProfile(leader).FullName())

			env.process(env.message(&leader, "/my"))
			my := env.waitSendMessage(leader.ID, event.Caption)
			require.Contains(t, my.String("text"), leaderLink)
			require.Contains(t, my.String("text"), followerLink)
			env.process(env.message(&leader, "/start "+signupPayload(event.ID, models.RoleLeader)))
			registeredScene := env.waitSendMessage(leader.ID, "Вы записаны в паре")
			require.Contains(t, registeredScene.String("text"), followerLink)

			stored := env.eventGet(event.ID)
			require.Empty(t, stored.Singles)
			require.Len(t, stored.Couples, 1)
			require.Equal(t, models.NewProfile(leader), *stored.Couples[0].Dancers[0].Profile)
			require.Equal(t, models.NewProfile(leader).FullName(), stored.Couples[0].Dancers[0].FullName)
			require.Equal(t, models.NewProfile(follower), *stored.Couples[0].Dancers[1].Profile)
			require.Equal(t, models.NewProfile(follower).FullName(), stored.Couples[0].Dancers[1].FullName)
			require.Equal(t, models.NewProfile(leader), stored.Owner)
		})
	}
}

func TestSignupIdenticalShortenedSingleNames(t *testing.T) {
	env := newEnv(t)
	first := models.Profile{ID: 101, FirstName: strings.Repeat("Я", 64), LastName: "First"}
	second := models.Profile{ID: 102, FirstName: strings.Repeat("Я", 64), LastName: "Second"}
	event := testEvent("identical_short_names", "Identical short names", userJohn)
	event.Singles = []models.Dancer{
		{Profile: &first, FullName: first.FullName(), Role: models.RoleFollower, AsSingle: true, CreatedAt: time.Now().UTC()},
		{Profile: &second, FullName: second.FullName(), Role: models.RoleFollower, AsSingle: true, CreatedAt: time.Now().UTC()},
	}
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	scene := env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
	shortName := strings.Repeat("Я", 63) + locale.NameEllipsis
	require.Equal(t, "1. "+shortName, scene.ReplyMarkup().Get("keyboard.1.0.text").String())
	secondCaption := scene.ReplyMarkup().Get("keyboard.2.0.text").String()
	require.Equal(t, "2. "+shortName, secondCaption)
	env.process(env.message(userJohn, secondCaption))
	post := env.tg.Wait("editMessageText")
	require.Contains(t, post.String("text"), `<a href="tg://user?id=102">`+shortName+`</a>`)
	result := env.waitSendMessage(userJohn.ID, "Вы зарегистрировались в паре")
	require.Contains(t, result.String("text"), `<a href="tg://user?id=102">`+shortName+`</a>`)
	env.waitSendMessage(second.ID, "зарегистрировался с тобой")
	stored := env.eventGet(event.ID)
	require.Len(t, stored.Couples, 1)
	require.Equal(t, second, *stored.Couples[0].Dancers[1].Profile)
	require.Equal(t, second.FullName(), stored.Couples[0].Dancers[1].FullName)
	require.Len(t, stored.Singles, 1)
	require.Equal(t, first, *stored.Singles[0].Profile)
	require.Equal(t, first.FullName(), stored.Singles[0].FullName)
}

func TestSignupManualNameLength(t *testing.T) {
	tests := []struct{ length, limit int }{{64, 64}, {65, 64}, {8, 8}, {9, 8}}
	for _, tt := range tests {
		name := strings.Repeat("🙂", tt.length)
		t.Run(fmt.Sprintf("limit=%d/length=%d", tt.limit, tt.length), func(t *testing.T) {
			env := newEnvWithConfig(t, func(cfg *config.Config) { cfg.DancerNameMaxLen = tt.limit })
			event := testEvent("manual_name_length", "Manual name length", userJohn)
			require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
			env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
			env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
			env.process(env.message(userJohn, name))
			if tt.length > tt.limit {
				env.waitSendMessage(userJohn.ID, locale.ErrDancerNameTooLong)
				require.Empty(t, env.eventGet(event.ID).Couples)
				return
			}
			post := env.tg.Wait("editMessageText")
			require.Contains(t, post.String("text"), name)
			result := env.waitSendMessage(userJohn.ID, "Вы зарегистрировались в паре")
			require.Contains(t, result.String("text"), name)
			require.Equal(t, name, env.eventGet(event.ID).Couples[0].Dancers[1].FullName)
		})
	}
}
