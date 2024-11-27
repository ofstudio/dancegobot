package views

import (
	"fmt"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
)

// AnswerQueryEmpty sends a response to the empty inline query.
func AnswerQueryEmpty(c tele.Context, thumb string) error {
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

// AnswerQuery sends a response to the non-empty inline query.
func AnswerQuery(c tele.Context, eventID, thumb string) error {
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
					ReplyMarkup: btnAnnouncement(eventID),
				},
				Title:       text,
				Description: desc,
				ThumbURL:    thumb,
				HideURL:     true,
			},
		},
	})
}
