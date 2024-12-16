package models

import (
	"log/slog"
	"time"
)

// HistoryItem - Event history item
type HistoryItem struct {
	Action    HistoryAction `json:"type"`                // Action of the action
	Initiator *Profile      `json:"initiator,omitempty"` // Telegram profile who initiated the action (if any)
	EventID   *string       `json:"event_id,omitempty"`  // ID of the event related to the action (if any)
	Details   any           `json:"details"`             // Payload of the action
	CreatedAt time.Time     `json:"created_at"`          // Creation time
}

// LogValue implements slog.Valuer interface for HistoryItem model.
func (h HistoryItem) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(h.Action)),
	}

	if h.Initiator != nil {
		attrs = append(attrs, slog.Int64("profile_id", h.Initiator.ID))
	}

	if h.EventID != nil {
		attrs = append(attrs, slog.Any("event", slog.GroupValue(
			slog.String("id", *h.EventID),
		)))
	}

	return slog.GroupValue(attrs...)
}

// HistoryAction - type of HistoryItem
type HistoryAction string

const (
	HistoryEventCreated     HistoryAction = "event_created"
	HistoryEventClosed      HistoryAction = "event_closed"
	HistoryEventReopened    HistoryAction = "event_reopened"
	HistoryCoupleAdded      HistoryAction = "couple_added"
	HistoryCoupleRemoved    HistoryAction = "couple_removed"
	HistorySingleAdded      HistoryAction = "single_added"
	HistorySingleRemoved    HistoryAction = "single_removed"
	HistoryNotificationSent HistoryAction = "notification_sent"
	HistoryPostAdded        HistoryAction = "post_added"
	HistoryPostChatAdded    HistoryAction = "post_chat_added"
)
