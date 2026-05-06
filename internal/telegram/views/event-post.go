package views

import (
	"fmt"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// EventPostAnswerEmpty sends a response to the empty inline query.
func EventPostAnswerEmpty(c tele.Context, thumb string) error {
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

// EventPostAnswer sends a response to the non-empty inline query.
func EventPostAnswer(c tele.Context, event *models.Event, thumb string) error {
	var desc string

	// Show warning in description if the text is too long.
	r := 255 - utf8.RuneCountInString(event.Caption)
	switch {
	case r < 0:
		desc = locale.QueryOverflow
	case r < 40:
		desc = fmt.Sprintf(locale.QueryRemaining, r, locale.NumQueryRemainingSymbols.N(r))
	case event.Settings.Limit > 0:
		desc = fmt.Sprintf(
			locale.QueryEventLimit,
			event.Settings.Limit,
			locale.NumLimitCouples.N(event.Settings.Limit),
		)
	default:
		desc = locale.QueryDescription
	}

	return c.Answer(&tele.QueryResponse{
		Results: tele.Results{
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: event.ID,
					Content: &tele.InputTextMessageContent{
						Text:           fmtCaption(event.Caption),
						ParseMode:      tele.ModeHTML,
						PreviewOptions: &tele.PreviewOptions{Disabled: true},
					},
					ReplyMarkup: btnEventSignupCb(event.ID),
				},
				Title:       event.Caption,
				Description: desc,
				ThumbURL:    thumb,
			},
		},
	})
}

var BtnEventSignupCb = tele.Btn{Unique: models.SessionSignup.String()}

// btnEventSignupCb creates signup buttons with callback data for the event post.
func btnEventSignupCb(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data(locale.RoleIcon[models.RoleLeader],
				BtnEventSignupCb.Unique,
				eventID, models.RoleLeader.String(),
				randtoken.New(4),
			),
			rm.Data(locale.RoleIcon[models.RoleFollower],
				BtnEventSignupCb.Unique,
				eventID, models.RoleFollower.String(),
				randtoken.New(4),
			),
		),
	)
	return rm
}
