package views

import (
	"errors"
	"html"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// Render returns services.RenderFunc function for services.RenderService
// that renders the event post with the given inline message ID.
func Render(api tele.API) func(*models.Event, string) error {
	return func(event *models.Event, inlineMessageID string) error {
		return render(api, event, inlineMessageID)
	}
}

// EventSignupURL returns a deeplink url for the event signup.
func EventSignupURL(eventID string, role models.Role) string {
	return telegram.Deeplink{Action: models.SessionSignup, EventID: eventID, Role: role}.String()
}

// render renders the event post with the given inline message ID.
func render(api tele.API, event *models.Event, inlineMessageID string) error {
	var rm *tele.ReplyMarkup
	if event.Settings.Closed {
		rm = btnEventClosed()
	} else {
		rm = btnEventSignupURL(event.ID)
	}

	msg := &tele.InlineResult{MessageID: inlineMessageID}
	opts := &tele.SendOptions{
		ReplyMarkup:           rm,
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}

	_, err := api.Edit(msg, postTextBuilder(event).String(), opts)
	if errors.Is(err, tele.ErrTrueResult) {
		return nil
	}
	return err
}

// postTextBuilder returns strings.Builder with the event post text.
func postTextBuilder(event *models.Event) *strings.Builder {
	sb := &strings.Builder{}
	if event.Settings.Closed {
		sb.WriteString(locale.IconPostClosed)
	}
	sb.WriteString(event.Caption)
	sb.WriteString("\n\n")

	if len(event.Couples) > 0 {
		sb.WriteString(locale.PostCouples)
		postCouplesBuild(sb, event.Couples, event.Settings.Limit)
		sb.WriteByte('\n')
	}

	if len(event.Singles) > 0 {
		leaders, followers := singlesByRole(event.Singles)
		if len(leaders) > len(followers) {
			sb.WriteString(locale.PostSingles[models.RoleLeader])
			postSinglesBuild(sb, leaders, followers)
		} else {
			sb.WriteString(locale.PostSingles[models.RoleFollower])
			postSinglesBuild(sb, followers, leaders)
		}
	}
	return sb
}

// postCouplesBuild appends the couples list to the strings.Builder.
// The limit parameter is used to separate the list into two parts: the second part is shown as a waitlist.
// The optional start parameter is used to set the index of the first couple.
func postCouplesBuild(sb *strings.Builder, couples []models.Couple, limit int, start ...int) {
	var startIndex int
	if len(start) > 0 {
		startIndex = start[0]
	}
	for i, c := range couples {
		if limit > 0 && i == limit {
			sb.WriteString(locale.PostCouplesWait)
		}
		sb.WriteString(strconv.Itoa(startIndex + i + 1))
		sb.WriteString(". ")
		sb.WriteString(fmtDancer(c.Dancers[0]))
		sb.WriteString(" – ")
		sb.WriteString(fmtDancer(c.Dancers[1]))
		sb.WriteByte('\n')
	}
}

// postSinglesBuild appends the singles lists s1 and s2 to the strings.Builder.
func postSinglesBuild(sb *strings.Builder, s1, s2 []models.Dancer) {
	for i, s := range s1 {
		singleBuild(sb, i+1, s)
	}
	if len(s1) > 0 && len(s2) > 0 {
		sb.WriteByte('\n')
	}
	for i, s := range s2 {
		singleBuild(sb, i+1, s)
	}
}

// singleBuild appends the single dancer to the strings.Builder.
func singleBuild(sb *strings.Builder, i int, single models.Dancer) {
	sb.WriteString(strconv.Itoa(i))
	sb.WriteString(". ")
	sb.WriteString(fmtDancer(single))
	sb.WriteByte('\n')
}

// singlesByRole separates singles by role. Returns leaders and followers slices.
func singlesByRole(singles []models.Dancer) ([]models.Dancer, []models.Dancer) {
	var leaders, followers []models.Dancer
	for _, d := range singles {
		if d.Role == models.RoleLeader {
			leaders = append(leaders, d)
		} else {
			followers = append(followers, d)
		}
	}
	return leaders, followers
}

var BtnEventClosed = tele.Btn{Unique: "post_closed"}

// btnEventClosed creates a button for the closed event post.
func btnEventClosed() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.Data(locale.IconPostClosed, BtnEventClosed.Unique, randtoken.New(4)),
	))
	return rm
}

// btnEventSignupURL creates signup buttons with url for the event post.
func btnEventSignupURL(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.URL(locale.RoleIcon[models.RoleLeader], EventSignupURL(eventID, models.RoleLeader)),
		rm.URL(locale.RoleIcon[models.RoleFollower], EventSignupURL(eventID, models.RoleFollower)),
	))
	return rm
}

// fmtDancer formats the dancer with a link to the Telegram profile.
func fmtDancer(d models.Dancer) string {
	name := html.EscapeString(d.FullName)
	if d.Profile == nil {
		return name
	}
	return `<a href="` + html.EscapeString(profileURL(d.Profile)) + `">` + name + "</a>"
}

// fmtProfile formats the profile with a link to Telegram profile.
func fmtProfile(p *models.Profile) string {
	if p == nil {
		return ""
	}
	return `<a href="` + html.EscapeString(profileURL(p)) + `">` + html.EscapeString(p.FullName()) + "</a>"
}

// profileURL formats the Telegram profile URL.
//
// If profile has a username, the link is created to the username.
// Example: https://t.me/username
//
// If the profile has no username, the link is created to the user ID.
// Example: tg://user?id=123456789
func profileURL(profile *models.Profile) string {
	if profile == nil {
		return ""
	}
	if profile.Username != "" {
		return "https://t.me/" + profile.Username
	}
	return "tg://user?id=" + strconv.FormatInt(profile.ID, 10)
}
