package models

import (
	"log/slog"
)

func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.Any("owner", e.Owner),
	)
}

func (d Dancer) LogValue() slog.Value {
	var attrs []slog.Attr
	if d.Profile != nil {
		attrs = append(attrs, slog.Any("profile_id", d.Profile.ID))
	}
	attrs = append(attrs,
		slog.String("full_name", d.FullName),
		slog.String("role", string(d.Role)),
		slog.String("status", d.Status.String()),
	)
	if d.SingleSignup {
		attrs = append(attrs, slog.Any("single_signup", d.SingleSignup))
	}
	return slog.GroupValue(attrs...)
}

func (p Profile) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Int64("id", p.ID),
		slog.String("first_name", p.FirstName),
	}
	if p.LastName != "" {
		attrs = append(attrs, slog.String("last_name", p.LastName))
	}
	if p.Username != "" {
		attrs = append(attrs, slog.String("username", p.Username))
	}
	return slog.GroupValue(attrs...)
}

func (h HistoryItem) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(h.Action)),
	}

	if h.Initiator != nil {
		attrs = append(attrs, slog.Int64("profile_id", h.Initiator.ID))
	}

	if h.EventID != nil {
		attrs = append(attrs, slog.String("event_id", *h.EventID))
	}

	return slog.GroupValue(attrs...)
}

func (u EventUpdate) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("result", u.Result.String()),
		slog.Any("event_id", u.Event.ID),
	}
	if u.Dancer != nil {
		attrs = append(attrs, slog.Any("dancer", u.Dancer))
	}
	if u.ChosenPartner != nil {
		attrs = append(attrs, slog.Any("chosen_partner", u.ChosenPartner))
	}
	return slog.GroupValue(attrs...)
}

func (n Notification) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("template", string(n.TmplCode)),
		slog.Any("recipient", n.Recipient),
	}
	if n.Initiator != nil {
		attrs = append(attrs, slog.Any("initiator", n.Initiator))
	}
	if n.Event != nil {
		attrs = append(attrs, slog.Any("event_id", n.Event.ID))
	}
	return slog.GroupValue(attrs...)
}
