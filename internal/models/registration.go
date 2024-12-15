package models

import (
	"fmt"
	"log/slog"
)

// Registration represents the registration of a dancer for an event.
type Registration struct {
	*Dancer                    // Dancer who is registered
	Status  RegistrationStatus // Current registration status for the event
	Result  RegistrationResult // The result of the registration request
	Event   *Event             // Event related to the registration
	Partner *Dancer            // Partner of the dancer if registered in a couple
	// Related contains other registration that is related by the current registration request (if any).
	//   - When a dancer tries to registers in a couple, Related contains the registration of the chosen partner.
	//   - When a dancer removes their registration in a couple, Related contains the registration of ex-partner.
	Related *Registration
}

// LogValue implements the slog.Valuer interface for Registration model.
func (r *Registration) LogValue() slog.Value {
	var attrs []slog.Attr
	if r.Event != nil {
		attrs = append(attrs, slog.Any("event", slog.GroupValue(
			slog.String("id", r.Event.ID),
		)))
	}
	attrs = append(attrs, r.attrs()...)
	if r.Related != nil {
		attrs = append(attrs, slog.Any("related", slog.GroupValue(
			r.Related.attrs()...,
		)))
	}

	return slog.GroupValue(attrs...)
}

func (r *Registration) attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("status", r.Status.String()),
	}
	if r.Result != ResultNoResult {
		attrs = append(attrs, slog.String("result", r.Result.String()))
	}
	attrs = append(attrs, slog.Any("", r.Dancer.LogValue()))
	if r.Partner != nil {
		attrs = append(attrs, slog.Any("partner", r.Partner.LogValue()))
	}
	return attrs
}

// RegistrationStatus is the status of the registration.
type RegistrationStatus int

const (
	StatusNotRegistered RegistrationStatus = iota // Not registered yet
	StatusAsSingle                                // Registered as a single
	StatusInCouple                                // Registered in a couple with a partner
	StatusForbidden                               // Forbidden to register for the event
)

// CanRegister returns true if the dancer can register for the event.
// The dancer can sign up if they are not registered yet or signed up as single.
func (s RegistrationStatus) CanRegister() bool {
	return s == StatusNotRegistered || s == StatusAsSingle
}

// IsRegistered returns true if the dancer is registered for the event as single or in a couple.
func (s RegistrationStatus) IsRegistered() bool {
	return s == StatusAsSingle || s == StatusInCouple
}

func (s RegistrationStatus) String() string {
	switch s {
	case StatusNotRegistered:
		return "not_registered"
	case StatusAsSingle:
		return "as_single"
	case StatusInCouple:
		return "in_couple"
	case StatusForbidden:
		return "forbidden"
	default:
		return fmt.Sprintf("unknown_status_%d", s)
	}
}

// RegistrationResult is the result of the registration request.
type RegistrationResult int

const (
	ResultNoResult            RegistrationResult = iota // No result
	ResultRegisteredAsSingle                            // Successful registration as single
	ResultRegisteredInCouple                            // Successful registration in a couple
	ResultRegistrationRemoved                           // Successful removal of registration
	ResultAlreadyAsSingle                               // The dancer is already registered as single
	ResultAlreadyInCouple                               // The dancer is already registered in another couple
	ResultAlreadyInSameCouple                           // The dancer is already registered in same couple
	ResultPartnerTaken                                  // Partner is already registered in another couple
	ResultPartnerSameRole                               // Partner has the same role as dancer
	ResultSelfNotAllowed                                // Not allowed to register in couple with yourself
	ResultWasNotRegistered                              // The dancer was not registered for the event
	ResultEventClosed                                   // The event is closed for new registrations
	ResultDancerForbidden                               // The event is forbidden for the dancer
	ResultPartnerForbidden                              // The event is forbidden for given partner
	ResultClosedForSingles                              // The event is closed for singles
	ResultClosedForSingleRole                           // The event is closed for singles  with given role
)

// IsSuccess returns true if the registration was successful.
func (r RegistrationResult) IsSuccess() bool {
	return r == ResultRegisteredAsSingle ||
		r == ResultRegisteredInCouple ||
		r == ResultRegistrationRemoved
}

// IsRetryable returns true if the registration can be retried with different parameters.
// Example: if the partner is taken, the registration can be retried with another partner.
func (r RegistrationResult) IsRetryable() bool {
	return r == ResultPartnerTaken ||
		r == ResultPartnerSameRole ||
		r == ResultSelfNotAllowed ||
		r == ResultPartnerForbidden ||
		r == ResultClosedForSingles ||
		r == ResultClosedForSingleRole
}

func (r RegistrationResult) String() string {
	switch r {
	case ResultNoResult:
		return "no_result"
	case ResultRegisteredAsSingle:
		return "registered_as_single"
	case ResultRegisteredInCouple:
		return "registered_in_couple"
	case ResultRegistrationRemoved:
		return "registration_removed"
	case ResultAlreadyAsSingle:
		return "already_as_single"
	case ResultAlreadyInCouple:
		return "already_in_couple"
	case ResultAlreadyInSameCouple:
		return "already_in_same_couple"
	case ResultPartnerTaken:
		return "partner_taken"
	case ResultPartnerSameRole:
		return "partner_same_role"
	case ResultSelfNotAllowed:
		return "self_not_allowed"
	case ResultWasNotRegistered:
		return "not_registered"
	case ResultEventClosed:
		return "event_closed"
	case ResultDancerForbidden:
		return "dancer_forbidden"
	case ResultPartnerForbidden:
		return "partner_forbidden"
	case ResultClosedForSingles:
		return "closed_for_singles"
	case ResultClosedForSingleRole:
		return "closed_for_single_role"
	default:
		return fmt.Sprintf("unknown_result_%d", r)
	}
}
