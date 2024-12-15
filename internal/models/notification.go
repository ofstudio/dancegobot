package models

import "log/slog"

type Notification struct {
	TmplCode  NotificationTmpl    `json:"template"`        // Template of the notification
	Recipient *Profile            `json:"recipient"`       // Receiver of the notification
	Payload   NotificationPayload `json:"-"`               // Payload of the notification
	Error     string              `json:"error,omitempty"` // Error message during notification sending (if any)
}

// LogValue implements slog.Valuer interface for Notification model.
func (n Notification) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("template", string(n.TmplCode)),
		slog.Any("recipient", n.Recipient.LogValue()),
	}
	if n.Payload.Event != nil {
		attrs = append(attrs, slog.Any("event", slog.GroupValue(
			slog.String("id", n.Payload.Event.ID),
		)))
	}
	return slog.GroupValue(attrs...)
}

// NotificationPayload contains the context of the notification.
type NotificationPayload struct {
	Event      *Event  // Event related to the notification (if any)
	Partner    *Dancer // Current partner of the recipient (if any)
	NewPartner *Dancer // New partner of the recipient (if any)
}

type NotificationTmpl string

func (t NotificationTmpl) String() string {
	return string(t)
}

const (
	// TmplRegisteredWithSingle - someone registered in couple with a single recipient
	TmplRegisteredWithSingle NotificationTmpl = "registered_with_single"

	// TmplCanceledWithSingle - someone who previously registered in couple with recipient
	// from the singles list canceled the registration.
	// The recipient will be returned back to the singles list.
	TmplCanceledWithSingle NotificationTmpl = "canceled_with_single"

	// TmplCanceledByPartner - chosen partner canceled registration
	TmplCanceledByPartner NotificationTmpl = "canceled_by_partner"

	// TmplAutoPairPartnerFound - partner has been found for the recipient
	TmplAutoPairPartnerFound NotificationTmpl = "auto_pair_partner_found"

	// TmplAutoPairPartnerChanged - partner has been canceled the registration
	// and new partner has been chosen.
	TmplAutoPairPartnerChanged NotificationTmpl = "auto_pair_partner_changed"
)
