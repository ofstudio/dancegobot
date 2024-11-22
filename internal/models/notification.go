package models

type Notification struct {
	Recipient Profile          `json:"recipient"` // Receiver of the notification
	TmplCode  NotificationTmpl `json:"template"`  // Template of the notification
	Initiator *Dancer          `json:"initiator"` // Who initiates the notification
	Event     *Event           `json:"-"`         // Event related to the notification (if any)
	// Virtual fields
	EventID *string `json:"event_id,omitempty"` // ID of the event related to the notification (if any)
	Error   string  `json:"error,omitempty"`    // Error message during notification sending (if any)
}

type NotificationTmpl string

const (
	// TmplRegisteredWithSingle - someone registered in couple with a single recipient
	TmplRegisteredWithSingle NotificationTmpl = "registered_with_single"

	// TmplCanceledWithSingle - someone who previously registered in couple with recipient
	// from the singles list canceled the registration.
	// The recipient will be returned back to the singles list.
	TmplCanceledWithSingle = "cancelled_with_single"

	// TmplCanceledByPartner - chosen partner canceled registration
	TmplCanceledByPartner = "cancelled_by_partner"
)
