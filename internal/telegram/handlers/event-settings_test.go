package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
)

func TestEventSettingsUnavailable(t *testing.T) {
	callbacks := []struct {
		name    string
		unique  string
		data    string
		handler func(*Handlers, tele.Context) error
	}{
		{"settings scene", views.BtnEventSettings.Unique, "missing|0|rand", (*Handlers).EventSettingsScene},
		{"auto pairing", views.BtnEventSettingsAutoPair.Unique, "missing|0|rand", (*Handlers).CbEventSettingsToggles},
		{"close registration", views.BtnEventSettingsClose.Unique, "missing|0|rand", (*Handlers).CbEventSettingsToggles},
		{"limit scene", views.BtnEventSettingsLimit.Unique, "missing|0|0|rand", (*Handlers).CbEventSettingsLimitScene},
		{"limit number", views.BtnEventSettingsLimitNum.Unique, "missing|2|0", (*Handlers).CbEventSettingsLimitNum},
	}
	for _, result := range []struct {
		name string
		err  error
	}{
		{"missing event", nil},
		{"read failure", errors.New("read failed")},
	} {
		t.Run(result.name, func(t *testing.T) {
			for _, cb := range callbacks {
				t.Run(cb.name, func(t *testing.T) {
					st := &eventSettingsErrorStore{err: result.err}
					cfg := config.Default().Settings
					h := NewHandlers(cfg,
						services.NewEventService(cfg, st, nil, nil),
						services.NewUserService(cfg, st), nil)
					c := &eventSettingsErrorContext{
						Context: tele.NewContext(nil, tele.Update{Callback: &tele.Callback{
							Unique: cb.unique,
							Data:   cb.data,
						}}),
					}
					user := &models.User{
						Profile: models.Profile{ID: 1},
						Session: models.Session{
							Action:  models.SessionSignup,
							EventID: "previous",
							Role:    models.RoleLeader,
						},
					}
					c.Set("user", user)

					require.NotPanics(t, func() {
						require.NoError(t, cb.handler(h, c))
					})
					require.Equal(t, []string{"missing"}, st.eventIDs)
					require.Equal(t, []models.Session{{}}, st.sessions)
					require.Equal(t, models.Session{}, user.Session)
					require.Equal(t, []string{locale.ErrSomethingWrong}, c.alerts)
				})
			}
		})
	}
}

func TestEventSettingsInvalidLimitParameters(t *testing.T) {
	cases := []struct {
		name    string
		unique  string
		values  []string
		handler func(*Handlers, tele.Context) error
	}{
		{"limit", views.BtnEventSettingsLimitNum.Unique,
			[]string{"-1", "21", "", "abc", "1.5", "999999999999999999999999999999"},
			(*Handlers).CbEventSettingsLimitNum},
		{"page", views.BtnEventSettingsLimit.Unique,
			[]string{"-1", "2", "", "abc", "1.5", "999999999999999999999999999999"},
			(*Handlers).CbEventSettingsLimitScene},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range tc.values {
				t.Run(value, func(t *testing.T) {
					st := &eventSettingsErrorStore{}
					cfg := config.Default().Settings
					h := NewHandlers(cfg,
						services.NewEventService(cfg, st, nil, nil),
						services.NewUserService(cfg, st), nil)
					c := &eventSettingsErrorContext{
						Context: tele.NewContext(nil, tele.Update{Callback: &tele.Callback{
							Unique: tc.unique,
							Data:   "event|" + value + "|0|rand",
						}}),
					}
					user := &models.User{
						Profile: models.Profile{ID: 1},
						Session: models.Session{Action: models.SessionSignup, EventID: "previous"},
					}
					c.Set("user", user)

					require.NotPanics(t, func() {
						require.NoError(t, tc.handler(h, c))
					})
					require.Empty(t, st.eventIDs)
					require.Equal(t, []models.Session{{}}, st.sessions)
					require.Equal(t, models.Session{}, user.Session)
					require.Equal(t, []string{locale.ErrSomethingWrong}, c.alerts)
				})
			}
		})
	}
}

// Unimplemented store methods fail the test if an error path attempts further work.
type eventSettingsErrorStore struct {
	store.Store
	err      error
	eventIDs []string
	sessions []models.Session
}

func (s *eventSettingsErrorStore) EventGet(_ context.Context, id string) (*models.Event, error) {
	s.eventIDs = append(s.eventIDs, id)
	return nil, s.err
}

func (s *eventSettingsErrorStore) UserUpdateSession(_ context.Context, user *models.User) error {
	s.sessions = append(s.sessions, user.Session)
	return nil
}

type eventSettingsErrorContext struct {
	tele.Context
	alerts []string
}

func (c *eventSettingsErrorContext) RespondAlert(text string) error {
	c.alerts = append(c.alerts, text)
	return nil
}
