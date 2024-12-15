package models

import (
	"fmt"
	"time"
)

// Event - is a dance event
type Event struct {
	ID        string        `json:"id"`         // Random string to identify the event
	Caption   string        `json:"caption"`    // Event caption
	Post      *Post         `json:"post"`       // Event post in a Telegram chat
	Settings  EventSettings `json:"settings"`   // Event settings
	Couples   []Couple      `json:"couples"`    // List of couples signed in
	Singles   []Dancer      `json:"singles"`    // List of singles signed in
	Owner     Profile       `json:"owner"`      // Telegram profile of the event owner
	CreatedAt time.Time     `json:"created_at"` // Creation time
}

// LogValue implements slog.Valuer interface for Event model.
func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.Any("owner", e.Owner.LogValue()),
	)
}

// EventSettings - is a settings for the event
type EventSettings struct {
	Limit               int       `json:"limit,omitempty"`                 // Maximum number of couples allowed to sign-in. Zero means no limit
	ClosedFor           ClosedFor `json:"closed_for,omitempty"`            // Is event closed for new signups or modifications
	DisableChooseSingle bool      `json:"disable_choose_single,omitempty"` // Disable choose specific single dancer from the wait list
}

type ClosedFor string

const (
	ClosedForAll             ClosedFor = "all"              // Closed for all. No modifications allowed
	ClosedForSingles         ClosedFor = "singles"          // Closed for singles
	ClosedForSingleLeaders   ClosedFor = "single_leaders"   // Closed for single leaders
	ClosedForSingleFollowers ClosedFor = "single_followers" // Closed for single followers
)
