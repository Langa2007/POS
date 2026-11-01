// Package events dispatches tenant-scoped outbox events to transactional subscribers.
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Langa2007/POS/backend/internal/kernel/events"
	"github.com/Langa2007/POS/backend/internal/kernel/tenancy"
	"github.com/Langa2007/POS/backend/internal/platform/db"
	"github.com/jackc/pgx/v5"
)

var ErrNoSubscriber = errors.New("no subscriber registered for event")

type Handler func(context.Context, pgx.Tx, events.Event) error

type Registry struct{ handlers map[string]Handler }

func NewRegistry() *Registry { return &Registry{handlers: make(map[string]Handler)} }
func (r *Registry) Register(eventType string, handler Handler) error {
	if eventType == "" || handler == nil {
		return errors.New("event type and handler are required")
	}
	if _, exists := r.handlers[eventType]; exists {
		return fmt.Errorf("subscriber already registered for %s", eventType)
	}
	r.handlers[eventType] = handler
	return nil
}

type Dispatcher struct {
	store    *db.Store
	registry *Registry
}

func NewDispatcher(store *db.Store, registry *Registry) *Dispatcher {
	return &Dispatcher{store: store, registry: registry}
}

// DispatchTenant handles up to limit events. Subscriber effects and delivery acknowledgement share a transaction,
// so database effects are retried after rollback and committed exactly once with the acknowledgement.
func (d *Dispatcher) DispatchTenant(ctx context.Context, tenantID string, limit int) (int, error) {
	if limit < 1 {
		return 0, errors.New("dispatch limit must be positive")
	}
	if tenantID == "" {
		return 0, tenancy.ErrMissingTenant
	}
	delivered := 0
	for delivered < limit {
		found := false
		var handlerErr error
		err := d.store.TenantTx(ctx, tenantID, func(tx pgx.Tx) error {
			var event events.Event
			var payload []byte
			err := tx.QueryRow(ctx, `SELECT id::text,tenant_id::text,event_type,payload,occurred_at FROM outbox WHERE tenant_id=$1 AND delivered_at IS NULL ORDER BY occurred_at,id LIMIT 1 FOR UPDATE SKIP LOCKED`, tenantID).Scan(&event.ID, &event.TenantID, &event.Type, &payload, &event.OccurredAt)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			found = true
			event.Payload = json.RawMessage(payload)
			handler, ok := d.registry.handlers[event.Type]
			if !ok {
				handlerErr = fmt.Errorf("%w: %s", ErrNoSubscriber, event.Type)
				_, err = tx.Exec(ctx, `UPDATE outbox SET attempts=attempts+1 WHERE id=$1`, event.ID)
				return err
			}
			if _, err = tx.Exec(ctx, `SAVEPOINT outbox_handler`); err != nil {
				return err
			}
			if handlerErr = handler(ctx, tx, event); handlerErr != nil {
				if _, err = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT outbox_handler`); err != nil {
					return err
				}
				_, err = tx.Exec(ctx, `UPDATE outbox SET attempts=attempts+1 WHERE id=$1`, event.ID)
				return err
			}
			_, err = tx.Exec(ctx, `UPDATE outbox SET delivered_at=now(),attempts=attempts+1 WHERE id=$1`, event.ID)
			return err
		})
		if err != nil {
			return delivered, err
		}
		if !found {
			return delivered, nil
		}
		if handlerErr != nil {
			return delivered, handlerErr
		}
		delivered++
	}
	return delivered, nil
}
