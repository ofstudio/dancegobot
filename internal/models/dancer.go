package models

import (
	"log/slog"
	"time"
)

// Dancer - is a dancer participating in the event
type Dancer struct {
	*Profile            // Telegram profile of the dancer (if available)
	FullName  string    `json:"full_name"`           // Name of the dancer
	Role      Role      `json:"role"`                // Role of the dancer
	AsSingle  bool      `json:"as_single,omitempty"` // If dancer was registered as single
	CreatedAt time.Time `json:"created_at"`          // Creation time
}

// LogValue implements the slog.Valuer interface for Dancer model.
func (d Dancer) LogValue() slog.Value {
	var attrs []slog.Attr
	if d.Profile != nil {
		attrs = append(attrs, slog.Any("profile", slog.GroupValue(
			slog.Int64("id", d.Profile.ID),
		)))
	}
	attrs = append(attrs,
		slog.String("full_name", d.FullName),
		slog.String("role", d.Role.String()),
	)
	if d.AsSingle {
		attrs = append(attrs, slog.Bool("as_single", d.AsSingle))
	}
	return slog.GroupValue(attrs...)
}
