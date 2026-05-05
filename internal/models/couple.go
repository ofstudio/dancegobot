package models

import "time"

// Couple - is a pair of dancers
type Couple struct {
	Dancers   []Dancer  `json:"dancers"`             // Dancers in the couple. Should be exactly 2. Leader is the first one.
	CreatedBy Profile   `json:"created_by"`          // Who created couple
	AutoPair  bool      `json:"auto_pair,omitempty"` // Couple was paired automatically
	CreatedAt time.Time `json:"created_at"`          // Creation time
}

// AffectedCouples contains information about the affected couples after the event settings limit changed.
type AffectedCouples struct {
	Couples   []Couple // Affected couples
	Position  int      // Start index of the affected couples
	Increased bool     // True if the limit was increased, false if decreased
}
