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
	attrs := []slog.Attr{
		slog.String("full_name", d.FullName),
		slog.String("role", string(d.Role)),
		slog.String("status", d.Status.String()),
		slog.Bool("single_signup", d.SingleSignup),
	}
	if d.Profile != nil {
		attrs = append(attrs, slog.Any("profile_id", d.Profile.ID))
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

	if h.Profile != nil {
		attrs = append(attrs, slog.Int64("profile_id", h.Profile.ID))
	}

	if h.EventID != nil {
		attrs = append(attrs, slog.String("event_id", *h.EventID))
	}

	return slog.GroupValue(attrs...)
}

func (u EventUpdate) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Any("event", u.Event),
		slog.String("result", u.Result.String()),
	}
	if u.Dancer != nil {
		attrs = append(attrs, slog.Any("dancer", u.Dancer))
	}
	if u.Couple != nil {
		attrs = append(attrs, slog.Any("couple", u.Couple))
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
