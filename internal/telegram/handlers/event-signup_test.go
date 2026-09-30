package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestSingleNumber(t *testing.T) {
	h := &Handlers{}

	tests := []struct {
		text   string
		want   int
		wantOK bool
	}{
		{text: "1", want: 1, wantOK: true},
		{text: "1.", want: 1, wantOK: true},
		{text: "1. Alice", want: 1, wantOK: true},
		{text: "01. Alice", want: 1, wantOK: true},
		{text: "1 Alice", wantOK: false},
		{text: "Alice 1.", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, gotOK := h.singleNumber(tt.text)
			require.Equal(t, tt.wantOK, gotOK)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFmtSinglesShortensNamesAndPreservesSelection(t *testing.T) {
	h := &Handlers{cfg: config.Default().Settings}
	first := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 64), LastName: "First", Username: "first_user"}
	second := models.Profile{ID: 43, FirstName: strings.Repeat("Я", 64), LastName: "Second"}
	singles := []models.Dancer{
		{Profile: &first, FullName: first.FullName(), Role: models.RoleFollower},
		{Profile: &second, FullName: second.FullName(), Role: models.RoleFollower},
		{FullName: "Manual", Role: models.RoleFollower},
		{Profile: &first, FullName: first.FullName(), Role: models.RoleLeader},
	}
	got := h.fmtSingles(singles, models.RoleFollower)
	require.Equal(t, []models.SessionSingle{
		{Profile: first, Caption: "1. " + strings.Repeat("Я", 63) + locale.NameEllipsis + " (@first_user)"},
		{Profile: second, Caption: "2. " + strings.Repeat("Я", 63) + locale.NameEllipsis},
	}, got)
	for i, single := range got {
		n, ok := h.singleNumber(single.Caption)
		require.True(t, ok)
		require.Equal(t, i+1, n)
		require.Equal(t, singles[i].Profile.ID, got[n-1].Profile.ID)
		require.Equal(t, singles[i].FullName, singles[i].Profile.FullName())
	}
	h.cfg.DancerNameMaxLen = 8
	short := h.fmtSingles(singles, models.RoleFollower)
	require.Equal(t, "1. "+strings.Repeat("Я", 7)+locale.NameEllipsis+" (@first_user)", short[0].Caption)
	require.Equal(t, first, short[0].Profile)
}
