// Package events defines durable domain-event envelopes and transactional outbox writes.
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidEvent = errors.New("event requires an ID, tenant, type, and valid JSON payload")

type Event struct {
	ID         string          `json:"id"`
	TenantID   string          `json:"tenant_id"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func New(tenantID, eventType string, payload any) (Event, error) {
	id, err := ids.NewUUIDv7()
	if err != nil {
		return Event{}, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("encode event payload: %w", err)
	}
	return Event{ID: id, TenantID: tenantID, Type: eventType, Payload: body, OccurredAt: time.Now().UTC()}, nil
}

// Write persists the event using the caller's transaction so the business change and event commit atomically.
func Write(ctx context.Context, tx pgx.Tx, event Event) error {
	if event.ID == "" || event.TenantID == "" || event.Type == "" || !json.Valid(event.Payload) {
		return ErrInvalidEvent
	}
	_, err := tx.Exec(ctx, `INSERT INTO outbox(id,tenant_id,event_type,payload,occurred_at) VALUES($1,$2,$3,$4,$5)`, event.ID, event.TenantID, event.Type, event.Payload, event.OccurredAt)
	return err
}
