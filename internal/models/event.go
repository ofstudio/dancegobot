package models

import (
	"fmt"
	"time"
)

// Event - is a dance event
type Event struct {
	ID        string    `json:"id"`               // Random string to identify the event
	Caption   string    `json:"text"`             // Announcement message text
	MessageID string    `json:"message_id"`       // Announcement inline_message_id
	Limit     int       `json:"limit,omitempty"`  // Maximum number of couples allowed to sign-in. Zero means no limit
	Closed    bool      `json:"closed,omitempty"` // Is event closed for new signups or modifications
	Couples   []Couple  `json:"couples"`          // List of couples signed in
	Singles   []Dancer  `json:"singles"`          // List of singles signed in
	Owner     Profile   `json:"owner"`            // Telegram profile of the event owner
	Chat      *Chat     `json:"chat,omitempty"`   // Chat where the event was created (only if bot is a member)
	CreatedAt time.Time `json:"created_at"`       // Creation time
}

// EventUpdate is the result of the event update request.
type EventUpdate struct {
	Event         *Event       // Updated event
	Result        UpdateResult // Result of the update
	Dancer        *Dancer      // Information about dancer (if any) with status and partner fields filled (if applicable)
	ChosenPartner *Dancer      // Information about chosen partner (if any)
	Couple        *Couple      // Information about couple related to the update (if any)
}

// UpdateResult is the result of the event update.
type UpdateResult int

const (
	ResultUnknown               UpdateResult = iota // Unknown result
	ResultSuccess                                   // Successful event handling
	ResultAlreadyAsSingle                           // The dancer is already registered as single
	ResultAlreadyInCouple                           // The dancer is already registered in another couple
	ResultAlreadyInSameCouple                       // The dancer is already registered in same couple
	ResultPartnerTaken                              // Requested partner is already registered in another couple
	ResultPartnerSameRole                           // Requested partner has the same role as dancer
	ResultSelfNotAllowed                            // Not allowed to register in couple with yourself
	ResultNotRegistered                             // The dancer is not registered for the event
	ResultEventClosed                               // The event is closed for new registrations
	ResultEventForbiddenDancer                      // The event is forbidden for the dancer
	ResultEventForbiddenPartner                     // The event is forbidden for given partner
	ResultSinglesNotAllowed                         // Not allowed to register as single
	ResultSinglesNotAllowedRole                     // Not allowed to register as single with given role
)

func (r UpdateResult) Retryable() bool {
	return r == ResultPartnerTaken ||
		r == ResultPartnerSameRole ||
		r == ResultSelfNotAllowed ||
		r == ResultEventForbiddenPartner ||
		r == ResultSinglesNotAllowed ||
		r == ResultSinglesNotAllowedRole
}

func (r UpdateResult) String() string {
	switch r {
	case ResultSuccess:
		return "success"
	case ResultAlreadyAsSingle:
		return "already registered as single"
	case ResultAlreadyInCouple:
		return "already registered in a couple"
	case ResultAlreadyInSameCouple:
		return "already registered in same couple"
	case ResultPartnerTaken:
		return "partner already taken"
	case ResultPartnerSameRole:
		return "partner with same role"
	case ResultSelfNotAllowed:
		return "not allowed to register with yourself"
	case ResultNotRegistered:
		return "is not registered"
	case ResultEventClosed:
		return "event closed for registrations"
	case ResultEventForbiddenDancer:
		return "event forbidden for dancer"
	case ResultEventForbiddenPartner:
		return "event forbidden for partner"
	case ResultSinglesNotAllowed:
		return "single registrations not allowed"
	case ResultSinglesNotAllowedRole:
		return "single registrations not allowed with given role"
	default:
		return fmt.Sprintf("unknown result: %d", r)
	}
}
