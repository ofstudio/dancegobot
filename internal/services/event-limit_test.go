package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
)

func TestLimitChangeNotifyCurrentRegistration(t *testing.T) {
	for _, tc := range []struct {
		name       string
		increased  bool
		change     func(*models.Event)
		recipients []int64
	}{
		{"decreased limit unchanged", false, func(*models.Event) {}, []int64{10, 11, 13}},
		{"active couple removed", false, func(e *models.Event) {
			e.Couples = e.Couples[1:]
		}, []int64{13}},
		{"limit removed after decrease", false, func(e *models.Event) {
			e.Settings.Limit = 0
		}, nil},
		{"limit decreased further", false, func(e *models.Event) {
			e.Settings.Limit = 1
		}, []int64{10, 11, 13}},
		{"limit increased after decrease", false, func(e *models.Event) {
			e.Settings.Limit = 3
		}, []int64{13}},
		{"increased limit unchanged", true, func(*models.Event) {}, []int64{10, 11, 13}},
		{"limit decreased after increase", true, func(e *models.Event) {
			e.Settings.Limit = 3
		}, []int64{10, 11}},
		{"limit removed after increase", true, func(e *models.Event) {
			e.Settings.Limit = 0
		}, []int64{10, 11, 13}},
		{"affected couples removed", false, func(e *models.Event) {
			e.Couples = e.Couples[:2]
		}, nil},
		{"same partners registered again", false, func(e *models.Event) {
			e.Couples[2].CreatedAt = e.Couples[2].CreatedAt.Add(time.Second)
		}, []int64{13}},
		{"same partners registered again after increase", true, func(e *models.Event) {
			e.Couples[2].CreatedAt = e.Couples[2].CreatedAt.Add(time.Second)
		}, []int64{13}},
		{"current participant data", false, func(e *models.Event) {
			e.Couples[2].Dancers[0].FullName = "Updated Eve"
			e.Couples[3].Dancers[1].Profile.FirstName = "Updated"
		}, []int64{10, 11, 13}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := anotherSampleEvent()
			oldLimit := 4
			snapshot.Settings.Limit = 2
			tmpl := models.TmplEventLimitDecreased
			if tc.increased {
				oldLimit = 2
				snapshot.Settings.Limit = 4
				tmpl = models.TmplEventLimitIncreased
			}
			affected := NewEventHandler(&snapshot).LimitChangeGetAffected(oldLimit)
			current := cloneEvent(&snapshot)
			tc.change(current)
			handler := NewEventHandler(current).LimitChangeNotifyAffected(affected)
			var recipients []int64
			for _, notification := range handler.Notifications() {
				recipients = append(recipients, notification.Recipient.ID)
				require.Equal(t, tmpl, notification.TmplCode)
				require.Same(t, current, notification.Payload.Event)
				for _, couple := range current.Couples {
					for i, dancer := range couple.Dancers {
						if dancer.Profile != nil && dancer.Profile.ID == notification.Recipient.ID {
							require.Equal(t, dancer.Profile, notification.Recipient)
							require.Equal(t, couple.Dancers[i^1], *notification.Payload.Partner)
						}
					}
				}
			}
			require.Equal(t, tc.recipients, recipients)
			require.Empty(t, handler.History())
		})
	}
}

func TestEventServiceLimitChangeUsesCurrentEvent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*models.Event)
	}{
		{"couple promoted after cancellation", func(e *models.Event) { e.Couples = e.Couples[1:] }},
		{"same partners registered again", func(e *models.Event) {
			e.Couples[2].CreatedAt = e.Couples[2].CreatedAt.Add(time.Second)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, st := newEventServiceTest(t)
			ctx := context.Background()
			snapshot := anotherSampleEvent()
			snapshot.Settings.Limit = 2
			require.NoError(t, st.EventUpsert(ctx, &snapshot))
			affected := s.LimitChangeGetAffected(&snapshot, 4)
			user := &models.User{Profile: snapshot.Owner, Session: models.Session{
				EventID: snapshot.ID, AffectedCouples: &affected,
			}}
			require.NoError(t, st.UserUpsertProfile(ctx, user))

			current := cloneEvent(&snapshot)
			current.Caption = "Current caption"
			current.Couples[3].Dancers[1].Profile.FirstName = "Current Hank"
			tc.change(current)
			require.NoError(t, st.EventUpsert(ctx, current))
			savedUser, err := st.UserGet(ctx, user.Profile.ID)
			require.NoError(t, err)
			require.NotNil(t, savedUser.Session.AffectedCouples)
			sent := make(chan *models.Notification, 4)
			s.notifier = NewNotifierService(s.cfg, st, func(n *models.Notification) error {
				sent <- n
				return nil
			})
			require.NoError(t, s.LimitChangeNotifyAffected(ctx, snapshot.ID, *savedUser.Session.AffectedCouples))
			select {
			case n := <-sent:
				require.Equal(t, int64(13), n.Recipient.ID)
				require.Equal(t, "Current Hank", n.Recipient.FirstName)
				require.Equal(t, current, n.Payload.Event)
				require.Equal(t, models.TmplEventLimitDecreased, n.TmplCode)
			case <-time.After(time.Second):
				t.Fatal("expected current waitlisted couple to be notified")
			}
			require.Eventually(t, func() bool {
				return historyActionCount(t, st, models.HistoryNotificationSent) == 1
			}, time.Second, 10*time.Millisecond)
			var data []byte
			require.NoError(t, st.DB().QueryRow("SELECT data FROM history WHERE action = ?",
				models.HistoryNotificationSent).Scan(&data))
			var history struct {
				Details models.Notification `json:"details"`
			}
			require.NoError(t, json.Unmarshal(data, &history))
			require.Equal(t, int64(13), history.Details.Recipient.ID)
			require.Equal(t, models.TmplEventLimitDecreased, history.Details.TmplCode)
			require.Empty(t, history.Details.Error)
			select {
			case n := <-sent:
				t.Fatalf("unexpected stale notification to %d", n.Recipient.ID)
			default:
			}
		})
	}
}
