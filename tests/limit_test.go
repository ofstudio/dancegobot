package tests

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/h2non/gock"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func (suite *AppTestSuite) TestEventLimitWaitlistSignup() {
	suite.Run("new couple goes to waitlist when limit is full", func() {
		event := testLimitEvent("limit_waitlist_signup", userJane, 1,
			testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
		)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		payload := "rand-signup-" + event.ID + "-leader"
		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					body.Get("text").String() == locale.SignupNotRegistered
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/start " + payload,
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, locale.PostCouplesWait) &&
					strings.Contains(text, "John") &&
					strings.Contains(text, "Manual Partner")
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				text := body.Get("text").String()
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(text, "Manual Partner") &&
					strings.Contains(text, locale.ResultCoupleWaitlist)
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "Manual Partner",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		reg, err := suite.app.EventService.RegistrationGet(context.Background(), event.ID,
			models.NewProfile(*userJohn), models.RoleLeader)
		suite.Require().NoError(err)
		suite.Equal(models.StatusInCouple, reg.Status)
		suite.True(reg.WaitList)
	})
}

func (suite *AppTestSuite) TestEventLimitWaitlistLeftAfterRemoval() {
	suite.Run("waitlist couple is notified after active couple removal", func() {
		event := testLimitEvent("limit_waitlist_left", userJohn, 1,
			testLimitCouple(userJohn, limitUserBob, userJohn, 0, false),
			testLimitCouple(userJane, limitUserCarol, userJane, time.Second, false),
		)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		payload := "rand-signup-" + event.ID + "-leader"
		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(body.Get("text").String(), "Bob")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/start " + payload,
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, "Jane Doe") &&
					!strings.Contains(text, locale.PostCouplesWait)
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					body.Get("text").String() == locale.ResultSuccessRemoved
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJane.ID &&
					strings.Contains(body.Get("text").String(), "вышли из списка ожидания")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   locale.BtnDancerRemove,
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) TestEventLimitChangeNotification() {
	suite.Run("owner can notify couples affected by increased limit", func() {
		event := testLimitEvent("limit_change_notify", userJohn, 1,
			testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
			testLimitCouple(userJane, limitUserCarol, userJane, time.Second, false),
		)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, "Jane Doe") &&
					!strings.Contains(text, locale.PostCouplesWait)
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("chat_id").Int() != userJohn.ID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, locale.EventSettingsCaption) &&
					strings.Contains(text, "Приходят первые 2 пары")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("chat_id").Int() != userJohn.ID {
					return false
				}
				text := body.Get("text").String()
				markup := telegramReplyMarkup(body)
				return strings.Contains(text, locale.LimitChangedIncreased) &&
					strings.Contains(text, "Jane Doe") &&
					strings.Contains(markup.Get("inline_keyboard").Raw, "lim_chg_ntf|"+event.ID)
			}).
			JSON(telegock.Result(tele.Message{ID: 101}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fevt_set_lim_num|" + event.ID + "|2|0",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.EditMessageMarkup).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					body.Get("message_id").Int() == 101
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					body.Get("text").String() == locale.LimitChangedNotified
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJane.ID &&
					strings.Contains(body.Get("text").String(), "увеличил лимит пар")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 101, Chat: privateChat(userJohn)},
				Data:    "\flim_chg_ntf|" + event.ID + "|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
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

var (
	limitTestBaseTime = time.Date(2024, 5, 6, 12, 0, 0, 0, time.UTC)
	limitUserAlice    = &tele.User{ID: 201, FirstName: "Alice"}
	limitUserBob      = &tele.User{ID: 202, FirstName: "Bob"}
	limitUserCarol    = &tele.User{ID: 203, FirstName: "Carol"}
)
