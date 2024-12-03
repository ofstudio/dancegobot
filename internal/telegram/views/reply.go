package views

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

// SendStart sends a welcome message.
func SendStart(c tele.Context) error {
	rm := btnTry()
	text := fmt.Sprintf(locale.Start, config.BotProfile().Username)
	return c.Send(text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// SendSignup sends a message on user signup request.
func SendSignup(c tele.Context, dancer *models.Dancer, singles []models.SessionSingle) error {
	opts := &tele.SendOptions{
		ReplyMarkup:           btnSignup(dancer, singles),
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}

	switch dancer.Status {
	case models.StatusNotRegistered:
		return c.Send(locale.SignupNotRegistered, opts)
	case models.StatusSingle:
		return c.Send(fmt.Sprintf(locale.SignupSingle, locale.IconSingle[dancer.Role]), opts)
	case models.StatusInCouple:
		return c.Send(fmt.Sprintf(locale.SignupInCouple, fmtDancerName(dancer.Partner)), opts)
	case models.StatusForbidden:
		return c.Send(locale.SignupForbidden, opts)
	default:
		return c.Send(locale.ErrSomethingWrong, tele.RemoveKeyboard)
	}
}

// SendResult sends a message on user signup result.
func SendResult(c tele.Context, upd *models.EventUpdate, singles []models.SessionSingle) error {

	successMsg := func(status models.DancerStatus) string {
		switch status {
		case models.StatusSingle:
			return fmt.Sprintf(locale.ResultSuccessSingle, locale.IconSingle[upd.Dancer.Role])
		case models.StatusInCouple:
			return fmt.Sprintf(locale.ResultSuccessCouple, fmtDancerName(upd.Dancer.Partner))
		default:
			return locale.ResultSuccessDeleted
		}
	}

	opts := &tele.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             tele.ModeHTML,
	}
	if upd.Result.Retryable() {
		opts.ReplyMarkup = btnSignup(upd.Dancer, singles)
	} else {
		opts.ReplyMarkup = &tele.ReplyMarkup{RemoveKeyboard: true}
	}

	switch upd.Result {
	case models.ResultSuccess:
		return c.Send(successMsg(upd.Dancer.Status), opts)
	case models.ResultAlreadyAsSingle:
		return c.Send(fmt.Sprintf(locale.ResultAlreadyAsSingle, locale.IconSingle[upd.Dancer.Role]), opts)
	case models.ResultAlreadyInCouple:
		return c.Send(fmt.Sprintf(locale.ResultAlreadyInCouple, fmtDancerName(upd.Dancer.Partner)), opts)
	case models.ResultAlreadyInSameCouple:
		return c.Send(locale.ResultAlreadyInSameCouple, opts)
	case models.ResultPartnerTaken:
		return c.Send(fmt.Sprintf(locale.ResultPartnerTaken, fmtDancerName(upd.ChosenPartner)), opts)
	case models.ResultPartnerSameRole:
		return c.Send(locale.ResultPartnerSameRole, opts)
	case models.ResultSelfNotAllowed:
		return c.Send(locale.ResultSelfNotAllowed, opts)
	case models.ResultNotRegistered:
		return c.Send(locale.ResultNotRegistered, opts)
	case models.ResultEventClosed:
		return c.Send(locale.ResultEventClosed, opts)
	case models.ResultEventForbiddenDancer:
		return c.Send(locale.ResultEventForbiddenDancer, opts)
	case models.ResultEventForbiddenPartner:
		return c.Send(locale.ResultEventForbiddenPartner, opts)
	case models.ResultSinglesNotAllowed:
		return c.Send(locale.ResultSinglesNotAllowed, opts)
	case models.ResultSinglesNotAllowedRole:
		return c.Send(locale.ResultSinglesNotAllowedRole, opts)
	default:
		return c.Send(locale.ErrSomethingWrong, tele.RemoveKeyboard)
	}
}

// SendCloseOK sends a message on user session close.
func SendCloseOK(c tele.Context) error {
	return c.Send(locale.CloseOK, tele.RemoveKeyboard)
}
