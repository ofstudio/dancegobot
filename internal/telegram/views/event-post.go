package views

import (
	"fmt"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

var BtnEventSignupCb = tele.Btn{Unique: models.SessionSignup.String()}

// btnEventSignupCb creates signup buttons with callback data for the event post.
func btnEventSignupCb(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data(locale.RoleIcon[models.RoleLeader],
				BtnEventSignupCb.Unique,
				eventID, models.RoleLeader.String(), randtoken.New(4)),
			rm.Data(locale.RoleIcon[models.RoleFollower],
				BtnEventSignupCb.Unique,
				eventID, models.RoleFollower.String(), randtoken.New(4)),
		),
	)
	return rm
}

// EventCreateAnswerEmpty sends a response to the empty inline query.
func EventCreateAnswerEmpty(c tele.Context, thumb string) error {
	return c.Answer(&tele.QueryResponse{
		Results: tele.Results{
			&tele.ArticleResult{
				Title:       locale.QueryTextEmpty,
				Text:        locale.QueryTextEmpty,
				Description: locale.QueryDescriptionEmpty,
				ThumbURL:    thumb,
			},
		},
	})
}

// EventCreateAnswer sends a response to the non-empty inline query.
func EventCreateAnswer(c tele.Context, eventID, thumb string) error {
	text := c.Query().Text
	var desc string

	// Show warning in description if the text is too long.
	r := 255 - utf8.RuneCountInString(text)
	switch {
	case r < 0:
		desc = locale.QueryOverflow
	case r < 40:
		desc = fmt.Sprintf(locale.QueryRemaining, r, locale.NumSymbols.N(r))
	default:
		desc = locale.QueryDescription
	}

	return c.Answer(&tele.QueryResponse{
		Results: tele.Results{
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: eventID,
					Content: &tele.InputTextMessageContent{
						Text:           text,
						ParseMode:      tele.ModeHTML,
						PreviewOptions: &tele.PreviewOptions{Disabled: true},
					},
					ReplyMarkup: btnEventSignupCb(eventID),
				},
				Title:       text,
				Description: desc,
				ThumbURL:    thumb,
				HideURL:     true,
			},
		},
	})
}
