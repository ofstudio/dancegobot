package models

import "time"

// Couple - is a pair of dancers
type Couple struct {
	Dancers   []Dancer  `json:"dancers"`    // Dancers in the couple. Should be exactly 2. Leader is the first one.
	CreatedBy Profile   `json:"created_by"` // Who created couple
	CreatedAt time.Time `json:"created_at"` // Creation time
}
