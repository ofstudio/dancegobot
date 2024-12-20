package telegram

import (
	"errors"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

// RenderPost renders the event post with the given inline message ID.
func RenderPost(api tele.API) func(*models.Event, string) error {
	return func(event *models.Event, inlineMessageID string) error {
		return render(api, event, inlineMessageID)
	}
}

// render renders the event post.
func render(api tele.API, event *models.Event, inlineMessageID string) error {
	textSB := renderText(event)
	rm := btnPostURL(event.ID)
	msg := &tele.InlineResult{MessageID: inlineMessageID}
	opts := &tele.SendOptions{
		ReplyMarkup:           rm,
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}
	_, err := api.Edit(msg, textSB.String(), opts)
	if errors.Is(err, tele.ErrTrueResult) {
		return nil
	}
	return err
}

func renderText(event *models.Event) *strings.Builder {
	sb := &strings.Builder{}
	sb.WriteString(event.Caption)
	sb.WriteString("\n\n")

	if len(event.Couples) > 0 {
		sb.WriteString(locale.PostCouples)
		sbCouples(sb, event.Couples)
		sb.WriteByte('\n')
	}

	if len(event.Singles) > 0 {
		leaders, followers := singlesByRole(event.Singles)
		if len(leaders) > len(followers) {
			sb.WriteString(locale.PostSingles[models.RoleLeader])
			sbSingles(sb, leaders, followers)
		} else {
			sb.WriteString(locale.PostSingles[models.RoleFollower])
			sbSingles(sb, followers, leaders)
		}
	}
	return sb
}

func sbCouples(sb *strings.Builder, couples []models.Couple) {
	for i, c := range couples {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(". ")
		sb.WriteString(fmtDancer(&c.Dancers[0]))
		sb.WriteString(" – ")
		sb.WriteString(fmtDancer(&c.Dancers[1]))
		sb.WriteByte('\n')
	}
}

func sbSingles(sb *strings.Builder, s1, s2 []models.Dancer) {
	for i, s := range s1 {
		sbSingle(sb, i+1, s)
	}
	if len(s1) > 0 && len(s2) > 0 {
		sb.WriteByte('\n')
	}
	for i, s := range s2 {
		sbSingle(sb, i+1, s)
	}
}

func sbSingle(sb *strings.Builder, i int, single models.Dancer) {
	sb.WriteString(strconv.Itoa(i))
	sb.WriteString(". ")
	sb.WriteString(fmtDancer(&single))
	sb.WriteByte('\n')
}

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
