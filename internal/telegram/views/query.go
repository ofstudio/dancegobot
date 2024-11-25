package views

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
)

// AnswerQueryEmpty sends a response to the empty inline query.
func AnswerQueryEmpty(c tele.Context, thumb string) error {
	return c.Answer(&tele.QueryResponse{
		Results: tele.Results{
			&tele.ArticleResult{
				Title:       locale.QueryTitleEmoji + " " + locale.QueryTextEmpty,
				Text:        locale.QueryTextEmpty,
				Description: locale.QueryDescriptionEmpty,
				ThumbURL:    thumb,
			},
		},
	})
}

// AnswerQuery sends a response to the non-empty inline query.
func AnswerQuery(c tele.Context, eventID, thumb string) error {
	return c.Answer(&tele.QueryResponse{
		Results: tele.Results{
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID:          eventID,
					ParseMode:   tele.ModeHTML,
					ReplyMarkup: btnAnnouncement(eventID),
				},
				Title:       locale.QueryTitleEmoji + " " + c.Query().Text,
				Text:        c.Query().Text,
				Description: locale.QueryDescription,
				ThumbURL:    thumb,
				HideURL:     true,
			},
		},
	})
}
