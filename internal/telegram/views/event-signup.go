package views

import (
	"fmt"
	"math/rand"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

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
		rows = append(rows, rm.Row(rm.Text(locale.BtnDancerRemove)))
	}

	// Add "close" button
	rows = append(rows, rm.Row(rm.Text(locale.BtnClose)))

	rm.Reply(rows...)
	return rm
}

// SendSignupScene sends a signup scene to the user.
func SendSignupScene(c tele.Context, reg *models.Registration, singles []models.SessionSingle) error {
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

// SendResult sends a message on user signup result.
func SendResult(c tele.Context, reg *models.Registration, singles []models.SessionSingle) error {
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
		msg := fmt.Sprintf(locale.ResultSuccessCouple, fmtDancer(reg.Partner))
		if reg.WaitList {
			msg += locale.ResultCoupleWaitlist
		}
		return c.Send(msg, opts)
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
	case models.ResultDancerForbidden:
		return c.Send(locale.ResultDancerForbidden, opts)
	case models.ResultPartnerForbidden:
		return c.Send(locale.ResultPartnerForbidden, opts)
	case models.ResultEventClosed:
		return c.Send(locale.ResultEventClosed, opts)
	case models.ResultEventRemoved:
		return c.Send(locale.ResultEventRemoved, opts)

	default:
		_ = c.Send(locale.ErrSomethingWrong, tele.RemoveKeyboard)
		return fmt.Errorf("unexpected registration result: '%s'", reg.Result.String())
	}
}

// SendCloseOK sends a message on user session close.
func SendCloseOK(c tele.Context) error {
	return c.Send(locale.Ok, tele.RemoveKeyboard)
}
