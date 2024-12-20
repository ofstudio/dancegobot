package telegram

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"unicode/utf8"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

// fmtProfileURL formats the Telegram profile URL.
//
// If profile has a username, the link is created to the username.
// Example: https://t.me/username
//
// If the profile has no username, the link is created to the user ID.
// Example: tg://user?id=123456789
func fmtProfileURL(profile *models.Profile) string {
	if profile == nil {
		return ""
	}
	if profile.Username != "" {
		return "https://t.me/" + profile.Username
	}
	return "tg://user?id=" + strconv.FormatInt(profile.ID, 10)
}

// fmtDancer formats the dancer with a link to the Telegram profile.
func fmtDancer(d *models.Dancer) string {
	if d.Profile == nil {
		return d.FullName
	}
	return "<a href='" + fmtProfileURL(d.Profile) + "'>" + d.FullName + "</a>"
}

// fmtSingles makes [models.SessionSingle] from the list of singles with given role.
// Returns the list of profiles with reply button captions.
// Caption format: "1. Full Name (@username)"
// or just "1. Full Name" if no Telegram username.
func fmtSingles(singles []models.Dancer, role models.Role) []models.SessionSingle {
	var s []models.SessionSingle
	for i, d := range singles {
		if d.Profile == nil {
			continue
		}
		if d.Role == role {
			caption := strconv.Itoa(i+1) + ". " + d.FullName
			if d.Profile.Username != "" {
				caption += " (@" + d.Profile.Username + ")"
			}
			s = append(s, models.SessionSingle{
				Caption: caption,
				Profile: *d.Profile,
			})

		}
	}
	return s
}

var reSingleCapt = regexp.MustCompile(`^\d+\. .+$`)

// isSingleCaption checks if the text is a single button caption.
func isSingleCaption(text string) bool {
	return reSingleCapt.MatchString(text)
}

// btnTry creates a button for the "Try" option on the start message.
func btnTry() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.Query(locale.BtnTry, " "),
	))
	return rm
}

var BtnCbSignup = tele.Btn{Unique: models.SessionSignup.String()}

// btnPostCb creates callback buttons for the event post.
func btnPostCb(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(
			rm.Data(locale.RoleIcon[models.RoleLeader],
				BtnCbSignup.Unique,
				eventID, models.RoleLeader.String(), randtoken.New(4)),
			rm.Data(locale.RoleIcon[models.RoleFollower],
				BtnCbSignup.Unique,
				eventID, models.RoleFollower.String(), randtoken.New(4)),
		),
	)
	return rm
}

// btnPostURL creates URL buttons for the event post.
func btnPostURL(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	dlLeader := Deeplink{Action: models.SessionSignup, EventID: eventID, Role: models.RoleLeader}
	dlFollower := Deeplink{Action: models.SessionSignup, EventID: eventID, Role: models.RoleFollower}
	rm.Inline(rm.Row(
		rm.URL(locale.RoleIcon[models.RoleLeader], dlLeader.String()),
		rm.URL(locale.RoleIcon[models.RoleFollower], dlFollower.String()),
	))
	return rm
}

// btnSignupScene creates buttons for the signup scene.
func btnSignupScene(reg *models.Registration, singles []models.SessionSingle) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		Placeholder:    locale.SignupPlaceholder,
	}
	var rows []tele.Row

	// if dancer can signup
	if reg.Status.CanRegister() {
		// add user sharing button
		rows = append(rows, rm.Row(
			rm.User(locale.BtnSignupContact, &tele.ReplyRecipient{
				ID:              rand.Int31(),
				Quantity:        1,
				Bot:             tele.Flag(false),
				RequestName:     tele.Flag(true),
				RequestUsername: tele.Flag(true),
			})))
		// add event singles with the opposite role if auto-pairing is off
		if !reg.Event.Settings.AutoPairing {
			for _, s := range singles {
				rows = append(rows, rm.Row(rm.Text(s.Caption)))
			}
		}
	}

	// add "signup as single" button if auto-pairing is on or no singles available
	if reg.Status == models.StatusNotRegistered &&
		(reg.Event.Settings.AutoPairing || len(singles) == 0) {
		rows = append(rows, rm.Row(rm.Text(locale.BtnAsSingle[reg.Role])))
	}

	// Add "remove" button if the dancer is already registered
	if reg.Status.IsRegistered() {
		rows = append(rows, rm.Row(rm.Text(locale.BtnRemove)))
	}

	// Add "close" button
	rows = append(rows, rm.Row(rm.Text(locale.BtnClose)))

	rm.Reply(rows...)
	return rm
}

// btnChatLink creates an inline button with a link to the chat.
//
// Known Telegram limitations:
//   - Only messages in supergroups or channels can be linked
//   - Supergroup or channel can be either public or private
//   - Bot should be a member of supergroup or an admin of the channel
//
// Link format:
//
//	https://t.me/c/{chat_link_id}/{message_id}
//
// Where {chat_link_id} = - {chat_id} - 1000000000000
//
// For example:
//
//	message_id:     1234
//	chat_id:       -1001234567890 (supergroup or channel)
//	chat_link_id:  -(-1001234567890) - 1000000000000 = 1234567890
//
// Which gives us the link: https://t.me/c/1234567890/1234
func btnChatLink(event *models.Event) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}

	if event == nil ||
		event.Post == nil ||
		event.Post.Chat == nil ||
		event.Post.ChatMessageID == 0 ||
		(event.Post.Chat.Type != models.ChatSuper && event.Post.Chat.Type != models.ChatChannel) {
		return rm
	}

	chatLinkId := -event.Post.Chat.ID - 1000000000000
	url := fmt.Sprintf("https://t.me/c/%d/%d", chatLinkId, event.Post.ChatMessageID)
	rm.Inline(rm.Row(
		rm.URL(locale.BtnChatLink, url),
	))
	return rm

}

var (
	BtnCbSettingsAutoPair = tele.Btn{Unique: "settings_auto_pair"}
	BtnCbSettingsHelp     = tele.Btn{Unique: "settings_help"}
	BtnCbSettingsBack     = tele.Btn{Unique: "settings_back"}
)

// btnSettingsScene creates buttons for the settings scene.
func btnSettingsScene(settings *models.UserSettings) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}
	rm.Inline(
		rm.Row(
			rm.Data(locale.BtnAutoPairing[settings.Event.AutoPairing],
				BtnCbSettingsAutoPair.Unique,
				randtoken.New(4)),
		),
		rm.Row(
			rm.Data(locale.BtnSettingsHelp, BtnCbSettingsHelp.Unique, randtoken.New(4)),
		),
	)
	return rm
}

// btnSettingsBack creates a button to return to the settings scene.
func btnSettingsBack() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{
		RemoveKeyboard: true,
	}
	rm.Inline(rm.Row(
		rm.Data(locale.BtnBack, BtnCbSettingsBack.Unique, randtoken.New(4)),
	))
	return rm
}

// sendStart sends a welcome message.
func sendStart(c tele.Context) error {
	rm := btnTry()
	text := fmt.Sprintf(locale.Start, config.BotProfile().Username)
	return c.Send(text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// sendSignupScene sends a signup scene to the user.
func sendSignupScene(c tele.Context, reg *models.Registration, singles []models.SessionSingle) error {
	opts := &tele.SendOptions{
		ReplyMarkup:           btnSignupScene(reg, singles),
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}

	switch reg.Status {
	case models.StatusNotRegistered:
		return c.Send(locale.SignupNotRegistered, opts)
	case models.StatusAsSingle:
		return c.Send(fmt.Sprintf(locale.SignupSingle, locale.IconSingle[reg.Role]), opts)
	case models.StatusInCouple:
		return c.Send(fmt.Sprintf(locale.SignupInCouple, fmtDancer(reg.Partner)), opts)
	case models.StatusForbidden:
		return c.Send(locale.SignupForbidden, opts)
	default:
		_ = c.Send(locale.ErrSomethingWrong, tele.RemoveKeyboard)
		return fmt.Errorf("unexpected registration status: '%s'", reg.Status.String())
	}
}

// sendResult sends a message on user signup result.
func sendResult(c tele.Context, reg *models.Registration, singles []models.SessionSingle) error {
	opts := &tele.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}
	if reg.Result.IsRetryable() {
		opts.ReplyMarkup = btnSignupScene(reg, singles)
	} else {
		opts.ReplyMarkup = &tele.ReplyMarkup{RemoveKeyboard: true}
	}

	switch reg.Result {
	case models.ResultRegisteredAsSingle:
		return c.Send(fmt.Sprintf(locale.ResultSuccessSingle, locale.IconSingle[reg.Role]), opts)
	case models.ResultRegisteredInCouple:
		return c.Send(fmt.Sprintf(locale.ResultSuccessCouple, fmtDancer(reg.Partner)), opts)
	case models.ResultRegistrationRemoved:
		return c.Send(locale.ResultSuccessRemoved, opts)
	case models.ResultAlreadyAsSingle:
		return c.Send(fmt.Sprintf(locale.ResultAlreadyAsSingle, locale.IconSingle[reg.Role]), opts)
	case models.ResultAlreadyInCouple:
		return c.Send(fmt.Sprintf(locale.ResultAlreadyInCouple, fmtDancer(reg.Partner)), opts)
	case models.ResultAlreadyInSameCouple:
		return c.Send(locale.ResultAlreadyInSameCouple, opts)
	case models.ResultPartnerTaken:
		return c.Send(fmt.Sprintf(locale.ResultPartnerTaken, fmtDancer(reg.Related.Dancer)), opts)
	case models.ResultPartnerSameRole:
		return c.Send(locale.ResultPartnerSameRole, opts)
	case models.ResultSelfNotAllowed:
		return c.Send(locale.ResultSelfNotAllowed, opts)
	case models.ResultWasNotRegistered:
		return c.Send(locale.ResultNotRegistered, opts)
	case models.ResultEventClosed:
		return c.Send(locale.ResultEventClosed, opts)
	case models.ResultDancerForbidden:
		return c.Send(locale.ResultDancerForbidden, opts)
	case models.ResultPartnerForbidden:
		return c.Send(locale.ResultPartnerForbidden, opts)
	case models.ResultClosedForSingles:
		return c.Send(locale.ResultClosedForSingles, opts)
	case models.ResultClosedForSingleRole:
		return c.Send(locale.ResultClosedForSingleRole, opts)
	default:
		_ = c.Send(locale.ErrSomethingWrong, tele.RemoveKeyboard)
		return fmt.Errorf("unexpected registration result: '%s'", reg.Result.String())
	}
}

// sendCloseOK sends a message on user session close.
func sendCloseOK(c tele.Context) error {
	return c.Send(locale.Ok, tele.RemoveKeyboard)
}

// msgSettingsScene returns a message with the user settings.
func msgSettingsScene(settings *models.UserSettings) (string, *tele.ReplyMarkup) {
	text := locale.SettingsCaption +
		locale.SettingsAutoPairing[settings.Event.AutoPairing]
	rm := btnSettingsScene(settings)
	return text, rm
}

// msgSettingsHelp returns a message with the user settings help.
func msgSettingsHelp() (string, *tele.ReplyMarkup) {
	return locale.SettingsHelp, btnSettingsBack()
}

// answerQueryEmpty sends a response to the empty inline query.
func answerQueryEmpty(c tele.Context, thumb string) error {
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

// answerQuery sends a response to the non-empty inline query.
func answerQuery(c tele.Context, eventID, thumb string) error {
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
					ReplyMarkup: btnPostCb(eventID),
				},
				Title:       text,
				Description: desc,
				ThumbURL:    thumb,
				HideURL:     true,
			},
		},
	})
}
