package models

// Session - is the current bot session of the user
type Session struct {
	Action          SessionAction    `json:"action,omitempty"`           // User action related to session
	EventID         string           `json:"event_id,omitempty"`         // Current event id related to the session
	Role            Role             `json:"event_role,omitempty"`       // Current role related to the session
	Singles         []SessionSingle  `json:"event_singles,omitempty"`    // List of singles available for signup with the current user role
	MyEvents        []string         `json:"my_events,omitempty"`        // List of event IDs for /my scene
	AffectedCouples *AffectedCouples `json:"affected_couples,omitempty"` // Affected couples after the event settings limit was changed
}

// SessionAction - is a user action related to the session
type SessionAction string

const (
	SessionNoAction  SessionAction = ""          // No action
	SessionSignup    SessionAction = "signup"    // Event signup action
	SessionSubscribe SessionAction = "subscribe" // New event subscription action
)

func (a SessionAction) String() string {
	return string(a)
}

// SessionSingle - is a single dancer profiles
// available for signup with the current user role
// associated with reply button caption.
type SessionSingle struct {
	Profile Profile `json:"profile"` // Profile of the single dancer
	Caption string  `json:"caption"` // Reply button caption
}
