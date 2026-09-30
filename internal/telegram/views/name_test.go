package views

import (
	"html"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

var testNameMaxLen = config.Default().DancerNameMaxLen

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty"},
		{name: "short", text: "Jane Doe", want: "Jane Doe"},
		{name: "63 runes", text: strings.Repeat("Я", 63), want: strings.Repeat("Я", 63)},
		{name: "64 runes", text: strings.Repeat("Я", 64), want: strings.Repeat("Я", 64)},
		{name: "65 runes", text: strings.Repeat("Я", 65), want: strings.Repeat("Я", 63) + locale.NameEllipsis},
		{name: "emoji", text: strings.Repeat("🙂", 65), want: strings.Repeat("🙂", 63) + locale.NameEllipsis},
		{name: "full profile", text: strings.Repeat("A", 40) + " " + strings.Repeat("B", 40), want: strings.Repeat("A", 40) + " " + strings.Repeat("B", 22) + locale.NameEllipsis},
		{name: "129 runes", text: strings.Repeat("Ж", 129), want: strings.Repeat("Ж", 63) + locale.NameEllipsis},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DisplayName(tt.text, testNameMaxLen)
			require.Equal(t, tt.want, got)
			require.True(t, utf8.ValidString(got))
			require.LessOrEqual(t, utf8.RuneCountInString(got), 64)
		})
	}
}

func TestDisplayNameConfiguredLimit(t *testing.T) {
	name := strings.Repeat("🙂", 9)
	require.Equal(t, strings.Repeat("🙂", 7)+locale.NameEllipsis, DisplayName(name, 8))
	require.Equal(t, name, DisplayName(name, 9))
	require.Equal(t, locale.NameEllipsis, DisplayName(name, 1))
	require.Empty(t, DisplayName(name, 0))
	require.Empty(t, DisplayName(name, -1))
}

func TestNameFormattingBeforeHTMLEscaping(t *testing.T) {
	for _, name := range []string{
		strings.Repeat("Я", 62) + "&<",
		strings.Repeat("Я", 62) + "&<>",
	} {
		for _, username := range []string{"", "long_name_user"} {
			profile := models.Profile{ID: 42, FirstName: name, Username: username}
			dancer := models.Dancer{Profile: &profile, FullName: name}
			url := "tg://user?id=42"
			if username != "" {
				url = "https://t.me/" + username
			}
			wantName := name
			if utf8.RuneCountInString(name) > 64 {
				wantName = strings.Repeat("Я", 62) + "&" + locale.NameEllipsis
			}
			want := `<a href="` + url + `">` + html.EscapeString(wantName) + `</a>`
			require.Equal(t, want, fmtDancer(dancer, testNameMaxLen))
			require.Equal(t, want, fmtProfile(&profile, testNameMaxLen))
			require.Equal(t, html.EscapeString(wantName), fmtDancer(models.Dancer{FullName: name}, testNameMaxLen))
			require.Equal(t, name, dancer.FullName)
			require.Equal(t, name, profile.FirstName)
		}
	}
}

func TestPostTextShortensCoupleAndSingleNames(t *testing.T) {
	profile := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("Ж", 40)}
	dancer := models.Dancer{Profile: &profile, FullName: profile.FullName(), Role: models.RoleLeader}
	partner := models.Dancer{FullName: strings.Repeat("🙂", 65), Role: models.RoleFollower}
	event := &models.Event{
		Caption: "Test Event",
		Couples: []models.Couple{{Dancers: []models.Dancer{dancer, partner}}},
		Singles: []models.Dancer{dancer},
	}
	text := postTextBuilder(event, testNameMaxLen).String()
	require.Equal(t, 2, strings.Count(text, fmtDancer(dancer, testNameMaxLen)))
	require.Contains(t, text, strings.Repeat("🙂", 63)+locale.NameEllipsis)
	require.NotContains(t, text, profile.FullName())
	require.NotContains(t, text, partner.FullName)
	require.Equal(t, profile.FullName(), event.Singles[0].FullName)
}

func TestNotificationTextShortensPartnerAndOwnerNames(t *testing.T) {
	owner := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("Ж", 40)}
	partner := models.Profile{ID: 43, FirstName: strings.Repeat("🙂", 64), LastName: "Surname"}
	n := &models.Notification{
		TmplCode: models.TmplEventLimitDecreased,
		Payload: models.NotificationPayload{
			Event:   &models.Event{Caption: "Test Event", Owner: owner},
			Partner: &models.Dancer{Profile: &partner, FullName: partner.FullName()},
		},
	}
	text, err := notifyTextBuilder(testNotifyTemplate, n)
	require.NoError(t, err)
	require.Contains(t, text.String(), `<a href="tg://user?id=42">`+strings.Repeat("Я", 40)+" "+strings.Repeat("Ж", 22)+locale.NameEllipsis+`</a>`)
	require.Contains(t, text.String(), `<a href="tg://user?id=43">`+strings.Repeat("🙂", 63)+locale.NameEllipsis+`</a>`)
	require.NotContains(t, text.String(), owner.FullName())
	require.NotContains(t, text.String(), partner.FullName())
	require.Equal(t, owner, n.Payload.Event.Owner)
	require.Equal(t, partner.FullName(), n.Payload.Partner.FullName)
	shortText, err := notifyTextBuilder(newNotifyTemplate(8), n)
	require.NoError(t, err)
	require.Contains(t, shortText.String(), `<a href="tg://user?id=42">`+strings.Repeat("Я", 7)+locale.NameEllipsis+`</a>`)
	require.Contains(t, shortText.String(), `<a href="tg://user?id=43">`+strings.Repeat("🙂", 7)+locale.NameEllipsis+`</a>`)
	unchangedText, err := notifyTextBuilder(testNotifyTemplate, n)
	require.NoError(t, err)
	require.Equal(t, text.String(), unchangedText.String())
}

func TestSignupMessagesUseConfiguredNameLimit(t *testing.T) {
	profile := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("Ж", 40)}
	partner := models.Dancer{Profile: &profile, FullName: profile.FullName(), Role: models.RoleFollower}
	reg := models.Registration{
		Dancer:  models.Dancer{Role: models.RoleLeader},
		Event:   &models.Event{ID: "event_id"},
		Status:  models.StatusInCouple,
		Partner: &partner,
		Related: &models.Registration{Dancer: partner},
	}
	ctx := &sendNameContext{}
	want := `<a href="tg://user?id=42">` + strings.Repeat("Я", 7) + locale.NameEllipsis + `</a>`
	require.NoError(t, SignupScene(ctx, reg, nil, 8))
	require.Contains(t, ctx.what, want)
	for _, result := range []models.RegistrationResult{
		models.ResultRegisteredInCouple, models.ResultAlreadyInCouple, models.ResultPartnerTaken,
	} {
		reg.Result = result
		require.NoError(t, SendResult(ctx, reg, nil, "", 8))
		require.Contains(t, ctx.what, want)
	}
	require.Equal(t, profile.FullName(), partner.FullName)
}

func TestLimitNoticeUsesConfiguredNameLimit(t *testing.T) {
	profile := models.Profile{ID: 42, FirstName: strings.Repeat("Я", 40), LastName: strings.Repeat("Ж", 40)}
	couple := models.Couple{Dancers: []models.Dancer{
		{Profile: &profile, FullName: profile.FullName(), Role: models.RoleLeader},
		{FullName: strings.Repeat("🙂", 65), Role: models.RoleFollower},
	}}
	ctx := &sendNameContext{}
	require.NoError(t, SendLimitChanged(ctx,
		&models.Event{ID: "event_id", Settings: models.EventSettings{Limit: 1}},
		models.AffectedCouples{Couples: []models.Couple{couple}, Position: 1}, 8))
	require.Contains(t, ctx.what, `<a href="tg://user?id=42">`+strings.Repeat("Я", 7)+locale.NameEllipsis+`</a>`)
	require.Contains(t, ctx.what, strings.Repeat("🙂", 7)+locale.NameEllipsis)
	require.Equal(t, profile.FullName(), couple.Dancers[0].FullName)
}

type sendNameContext struct {
	tele.Context
	what interface{}
}

func (c *sendNameContext) Send(what interface{}, opts ...interface{}) error {
	c.what = what
	return nil
}
