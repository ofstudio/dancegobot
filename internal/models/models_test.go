package models

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventLogValue(t *testing.T) {
	event := Event{
		ID:                  "event_id",
		Owner:               Profile{ID: 42, FirstName: "John"},
		SubscribersNotified: true,
	}

	require.Equal(t, []slog.Attr{
		slog.String("id", "event_id"),
		slog.Any("owner", event.Owner.LogValue()),
		slog.Bool("subscribers_notified", true),
	}, event.LogValue().Group())
}

func TestProfileFullName(t *testing.T) {
	require.Equal(t, "John Doe", Profile{FirstName: "John", LastName: "Doe"}.FullName())
	require.Equal(t, "John", Profile{FirstName: "John"}.FullName())
}

func TestRoleOpposite(t *testing.T) {
	require.Equal(t, RoleFollower, RoleLeader.Opposite())
	require.Equal(t, RoleLeader, RoleFollower.Opposite())
}

func TestRegistrationStatus(t *testing.T) {
	tests := []struct {
		status       RegistrationStatus
		canRegister  bool
		isRegistered bool
		text         string
	}{
		{StatusNotRegistered, true, false, "not_registered"},
		{StatusAsSingle, true, true, "as_single"},
		{StatusInCouple, false, true, "in_couple"},
		{StatusForbidden, false, false, "forbidden"},
		{RegistrationStatus(99), false, false, "unknown_status_99"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			require.Equal(t, tt.canRegister, tt.status.CanRegister())
			require.Equal(t, tt.isRegistered, tt.status.IsRegistered())
			require.Equal(t, tt.text, tt.status.String())
		})
	}
}

func TestRegistrationResult(t *testing.T) {
	tests := []struct {
		result    RegistrationResult
		success   bool
		retryable bool
		text      string
	}{
		{ResultNoResult, false, false, "no_result"},
		{ResultRegisteredAsSingle, true, false, "registered_as_single"},
		{ResultRegisteredInCouple, true, false, "registered_in_couple"},
		{ResultRegistrationRemoved, true, false, "registration_removed"},
		{ResultAlreadyAsSingle, false, false, "already_as_single"},
		{ResultAlreadyInCouple, false, false, "already_in_couple"},
		{ResultAlreadyInSameCouple, false, false, "already_in_same_couple"},
		{ResultPartnerTaken, false, true, "partner_taken"},
		{ResultPartnerSameRole, false, true, "partner_same_role"},
		{ResultSelfNotAllowed, false, true, "self_not_allowed"},
		{ResultWasNotRegistered, false, false, "was_not_registered"},
		{ResultDancerForbidden, false, false, "dancer_forbidden"},
		{ResultPartnerForbidden, false, true, "partner_forbidden"},
		{ResultEventClosed, false, false, "event_closed"},
		{ResultEventRemoved, false, false, "event_removed"},
		{RegistrationResult(99), false, false, "unknown_result_99"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			require.Equal(t, tt.success, tt.result.IsSuccess())
			require.Equal(t, tt.retryable, tt.result.IsRetryable())
			require.Equal(t, tt.text, tt.result.String())
		})
	}
}
