package models

import "time"

// HistoryItem - Event history item
type HistoryItem struct {
	Action    HistoryAction `json:"type"`               // Action of the action
	Profile   *Profile      `json:"profile,omitempty"`  // Telegram profile related to the action
	EventID   *string       `json:"event_id,omitempty"` // ID of the event related to the action
	Details   any           `json:"details"`            // Payload of the action
	CreatedAt time.Time     `json:"created_at"`         // Creation time
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
)
