package event

import (
	"encoding/json"
	"fmt"
	"time"
)

type NotificationEvent struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewNotificationEvent(eventType string, payload interface{}) *NotificationEvent {
	return &NotificationEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// FormatSSE formata o evento conforme a especificação oficial de Server-Sent Events
func (e *NotificationEvent) FormatSSE() ([]byte, error) {
	data, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, err
	}
	formatted := fmt.Sprintf("event: %s\ndata: %s\n\n", e.Type, string(data))
	return []byte(formatted), nil
}

