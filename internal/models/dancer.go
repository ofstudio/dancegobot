package models

import (
	"fmt"
	"time"
)

// Dancer - is a dancer participating in the event
type Dancer struct {
	*Profile               // Telegram profile of the dancer (if available)
	FullName     string    `json:"full_name"`     // Name of the dancer
	Role         Role      `json:"role"`          // Role of the dancer
	SingleSignup bool      `json:"single_signup"` // Signed up as single
	CreatedAt    time.Time `json:"created_at"`    // Creation time
	// Virtual fields
	Status  DancerStatus `json:"-"` // Dancer status at the event. Not stored in the database.
	Partner *Dancer      `json:"-"` // Partner of the dancer if any. Not stored in the database.
}

type DancerStatus int

const (
	StatusNotRegistered DancerStatus = iota // Not registered yet
	StatusSingle                            // Registered as a single
	StatusInCouple                          // Registered in a couple with a partner
	StatusForbidden                         // Forbidden to register for the event
)

// SignupAvailable returns true if the dancer can sign up for the event.
// The dancer can sign up if they are not registered yet or signed up as single.
func (s DancerStatus) SignupAvailable() bool {
	return s == StatusNotRegistered || s == StatusSingle
}

// SignedUp returns true if the dancer is signed up for the event as single or in a couple.
func (s DancerStatus) SignedUp() bool {
	return s == StatusSingle || s == StatusInCouple
}

func (s DancerStatus) String() string {
	switch s {
	case StatusNotRegistered:
		return "not registered"
	case StatusSingle:
		return "single"
	case StatusInCouple:
		return "in couple"
	case StatusForbidden:
		return "forbidden"
	default:
		return fmt.Sprintf("unknown status: %d", s)
	}
}
