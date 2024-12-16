package models

// Session - is the current bot session of the user
type Session struct {
	Action  SessionAction   `json:"action,omitempty"`        // User action related to session
	EventID string          `json:"event_id,omitempty"`      // Current event id related to the session (if any)
	Role    Role            `json:"event_role,omitempty"`    // Current role related to the session (if any)
	Singles []SessionSingle `json:"event_singles,omitempty"` // Singles - list of singles available for signup with the current user role
}

// SessionAction - is a user action related to the session
type SessionAction string

const (
	SessionNoAction SessionAction = ""
	SessionSignup   SessionAction = "signup"
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
