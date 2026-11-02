package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalid = errors.New("invalid workflow")

type Store struct {
	pool interface {
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	}
}

func New(pool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}) *Store {
	return &Store{pool: pool}
}

type Workflow struct {
	ID         string          `json:"id"`
	Key        string          `json:"key"`
	Name       string          `json:"name"`
	Definition json.RawMessage `json:"definition"`
	Enabled    bool            `json:"enabled"`
}

type Transition struct {
	From     string          `json:"from"`
	To       string          `json:"to"`
	Requires json.RawMessage `json:"requires"`
	Actions  json.RawMessage `json:"actions"`
}

type Definition struct {
	States      []string     `json:"states"`
	Initial     string       `json:"initial"`
	Transitions []Transition `json:"transitions"`
}

func (s *Store) Upsert(ctx context.Context, tx pgx.Tx, tenantID string, w Workflow) (Workflow, error) {
	if strings.TrimSpace(w.Key) == "" || strings.TrimSpace(w.Name) == "" || len(w.Definition) == 0 {
		return Workflow{}, ErrInvalid
	}
	var d Definition
	if err := json.Unmarshal(w.Definition, &d); err != nil {
		return Workflow{}, ErrInvalid
	}
	if len(d.States) == 0 {
		return Workflow{}, ErrInvalid
	}
	id := w.ID
	if id == "" {
		nid, err := ids.NewUUIDv7()
		if err != nil {
			return Workflow{}, err
		}
		id = nid
	}
	_, err := tx.Exec(ctx, `INSERT INTO workflows(id,tenant_id,key,name,definition,enabled) VALUES($1,$2,$3,$4,$5,$6)
ON CONFLICT (tenant_id,key) DO UPDATE SET name=EXCLUDED.name, definition=EXCLUDED.definition, enabled=EXCLUDED.enabled, updated_at=now()`,
		id, tenantID, w.Key, w.Name, w.Definition, w.Enabled)
	if err != nil {
		return Workflow{}, err
	}
	w.ID = id
	return w, nil
}

func (s *Store) Get(ctx context.Context, tenantID, key string) (Workflow, error) {
	var w Workflow
	if err := s.pool.QueryRow(ctx, `SELECT id,key,name,definition,enabled FROM workflows WHERE tenant_id=$1 AND key=$2`, tenantID, key).Scan(&w.ID, &w.Key, &w.Name, &w.Definition, &w.Enabled); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Workflow{}, ErrInvalid
		}
		return Workflow{}, err
	}
	return w, nil
}

func CanTransition(d Definition, from, to string) bool {
	for _, t := range d.Transitions {
		if t.From == from && t.To == to {
			return true
		}
	}
	return false
}
