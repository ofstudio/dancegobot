package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestEventServiceDancerNameValidation(t *testing.T) {
	profile := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("🙂", 40)}
	tests := []struct {
		name    string
		dancer  models.Dancer
		limit   int
		wantErr string
	}{
		{name: "manual one rune", dancer: models.Dancer{FullName: "Я", Role: models.RoleLeader}},
		{name: "manual 64 runes", dancer: models.Dancer{FullName: strings.Repeat("🙂", 64), Role: models.RoleFollower}},
		{name: "manual 65 runes", dancer: models.Dancer{FullName: strings.Repeat("Я", 65), Role: models.RoleLeader}, wantErr: "full name must be between"},
		{name: "manual empty", dancer: models.Dancer{Role: models.RoleLeader}, wantErr: "full name must be between"},
		{name: "manual whitespace", dancer: models.Dancer{FullName: " \t ", Role: models.RoleLeader}, wantErr: "full name must be between"},
		{name: "manual trims whitespace", dancer: models.Dancer{FullName: " Я ", Role: models.RoleLeader}, limit: 1},
		{name: "manual custom limit", dancer: models.Dancer{FullName: "Яна", Role: models.RoleLeader}, limit: 2, wantErr: "full name must be between"},
		{name: "profile 81 runes", dancer: models.Dancer{Profile: &profile, FullName: profile.FullName(), Role: models.RoleLeader}},
		{name: "profile ignores manual limit", dancer: models.Dancer{Profile: &profile, FullName: profile.FullName(), Role: models.RoleFollower}, limit: 2},
		{name: "profile invalid ID", dancer: models.Dancer{Profile: &models.Profile{FirstName: "Яна"}, FullName: "Яна", Role: models.RoleLeader}, wantErr: "profile ID must be positive"},
		{name: "profile missing first name", dancer: models.Dancer{Profile: &models.Profile{ID: 42, LastName: "Surname"}, FullName: "Surname", Role: models.RoleLeader}, wantErr: "profile first name must be provided"},
		{name: "profile invalid role", dancer: models.Dancer{Profile: &profile, FullName: profile.FullName()}, wantErr: "invalid role"},
		{name: "manual invalid role", dancer: models.Dancer{FullName: "Яна"}, wantErr: "invalid role"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default().Settings
			if tt.limit != 0 {
				cfg.DancerNameMaxLen = tt.limit
			}
			service := &EventService{cfg: cfg}
			err := service.validateDancer(tt.dancer)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestEventServiceLongProfileNames(t *testing.T) {
	ctx := context.Background()
	dancer := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("🙂", 40)}
	partner := models.Profile{ID: 43, FirstName: strings.Repeat("Ж", 64), LastName: strings.Repeat("Ю", 64)}
	for _, scenario := range []string{"single", "manual partner", "profile partner", "single partner"} {
		t.Run(scenario, func(t *testing.T) {
			service, st := newEventServiceTest(t)
			event := &models.Event{ID: "long_names", Caption: "abc", Owner: models.Profile{ID: 1, FirstName: "Owner"}}
			require.NoError(t, st.EventUpsert(ctx, event))
			if scenario == "single partner" {
				reg, err := service.SingleAdd(ctx, event.ID, partner, models.RoleFollower)
				require.NoError(t, err)
				require.Equal(t, models.ResultRegisteredAsSingle, reg.Result)
			}

			var reg models.Registration
			var err error
			switch scenario {
			case "single":
				reg, err = service.SingleAdd(ctx, event.ID, dancer, models.RoleLeader)
			case "manual partner":
				reg, err = service.CoupleAdd(ctx, event.ID, &dancer, models.RoleLeader, "  Manual Partner  ")
			default:
				reg, err = service.CoupleAdd(ctx, event.ID, &dancer, models.RoleLeader, &partner)
			}
			require.NoError(t, err)
			require.Equal(t, dancer.FullName(), reg.FullName)
			require.Equal(t, dancer, *reg.Profile)

			stored, err := st.EventGet(ctx, event.ID)
			require.NoError(t, err)
			action := models.HistoryCoupleAdded
			if scenario == "single" {
				require.Equal(t, models.ResultRegisteredAsSingle, reg.Result)
				require.Len(t, stored.Singles, 1)
				require.Equal(t, dancer.FullName(), stored.Singles[0].FullName)
				require.Equal(t, dancer, *stored.Singles[0].Profile)
				action = models.HistorySingleAdded
			} else {
				require.Equal(t, models.ResultRegisteredInCouple, reg.Result)
				require.Len(t, stored.Couples, 1)
				require.Empty(t, stored.Singles)
				require.Equal(t, dancer.FullName(), stored.Couples[0].Dancers[0].FullName)
				require.Equal(t, dancer, *stored.Couples[0].Dancers[0].Profile)
				if scenario == "manual partner" {
					require.Equal(t, "Manual Partner", stored.Couples[0].Dancers[1].FullName)
					require.Nil(t, stored.Couples[0].Dancers[1].Profile)
				} else {
					require.Equal(t, partner.FullName(), stored.Couples[0].Dancers[1].FullName)
					require.Equal(t, partner, *stored.Couples[0].Dancers[1].Profile)
				}
			}
			require.Eventually(t, func() bool {
				return historyActionCount(t, st, action) == 1
			}, time.Second, time.Millisecond)
			if scenario == "single partner" {
				require.Eventually(t, func() bool {
					return historyActionCount(t, st, models.HistorySingleAdded) == 1 &&
						historyActionCount(t, st, models.HistorySingleRemoved) == 1 &&
						historyActionCount(t, st, models.HistoryNotificationSent) == 1
				}, time.Second, time.Millisecond)
			}
		})
	}
}

func TestEventServiceRejectsLongManualName(t *testing.T) {
	service, st := newEventServiceTest(t)
	ctx := context.Background()
	event := &models.Event{ID: "manual_name", Caption: "abc", Owner: models.Profile{ID: 1, FirstName: "Owner"}}
	require.NoError(t, st.EventUpsert(ctx, event))
	_, err := service.CoupleAdd(ctx, event.ID, &models.Profile{ID: 42, FirstName: "Dancer"},
		models.RoleLeader, strings.Repeat("Я", 65))
	require.ErrorContains(t, err, "full name must be between")
	stored, err := st.EventGet(ctx, event.ID)
	require.NoError(t, err)
	require.Equal(t, event, stored)
	require.Zero(t, historyActionCount(t, st, models.HistoryCoupleAdded))
	require.Zero(t, historyActionCount(t, st, models.HistoryNotificationSent))
}
