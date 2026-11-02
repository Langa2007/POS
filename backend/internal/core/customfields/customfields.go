package customfields

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalid = errors.New("invalid custom field")

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

type Def struct {
	ID         string          `json:"id"`
	EntityType string          `json:"entity_type"`
	FieldKey   string          `json:"field_key"`
	FieldType  string          `json:"field_type"`
	Label      string          `json:"label"`
	Options    json.RawMessage `json:"options"`
	Validation json.RawMessage `json:"validation"`
	Required   bool            `json:"required"`
	Searchable bool            `json:"searchable"`
	Active     bool            `json:"active"`
}

func allowed(t string) bool {
	switch t {
	case "text", "number", "decimal", "boolean", "date", "select", "multi-select", "file", "reference":
		return true
	}
	return false
}

func (s *Store) Upsert(ctx context.Context, tx pgx.Tx, tenantID string, d Def) (Def, error) {
	if d.EntityType == "" || d.FieldKey == "" || d.Label == "" {
		return Def{}, ErrInvalid
	}
	if !allowed(d.FieldType) {
		return Def{}, ErrInvalid
	}
	id := d.ID
	if id == "" {
		nid, err := ids.NewUUIDv7()
		if err != nil {
			return Def{}, err
		}
		id = nid
	}
	if d.Options == nil {
		d.Options = json.RawMessage("[]")
	}
	if d.Validation == nil {
		d.Validation = json.RawMessage("{}")
	}
	_, err := tx.Exec(ctx, `INSERT INTO custom_field_defs(id,tenant_id,entity_type,field_key,field_type,label,options,validation,required,searchable,active) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (tenant_id,entity_type,field_key) DO UPDATE SET label=EXCLUDED.label, field_type=EXCLUDED.field_type, options=EXCLUDED.options, validation=EXCLUDED.validation, required=EXCLUDED.required, searchable=EXCLUDED.searchable, active=EXCLUDED.active, updated_at=now()`,
		id, tenantID, strings.ToLower(d.EntityType), strings.ToLower(d.FieldKey), d.FieldType, d.Label, d.Options, d.Validation, d.Required, d.Searchable, d.Active)
	if err != nil {
		return Def{}, err
	}
	d.ID = id
	return d, nil
}

func (s *Store) List(ctx context.Context, tenantID, entityType string) ([]Def, error) {
	return s.list(ctx, s.pool, tenantID, entityType)
}

func (s *Store) ListTx(ctx context.Context, tx pgx.Tx, tenantID, entityType string) ([]Def, error) {
	return s.list(ctx, tx, tenantID, entityType)
}

func (s *Store) list(ctx context.Context, queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, tenantID, entityType string) ([]Def, error) {
	q := `SELECT id,entity_type,field_key,field_type,label,options,validation,required,searchable,active FROM custom_field_defs WHERE tenant_id=$1`
	args := []any{tenantID}
	if entityType != "" {
		q += ` AND entity_type=$2 ORDER BY created_at`
		args = append(args, strings.ToLower(entityType))
	} else {
		q += ` ORDER BY entity_type,created_at`
	}
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Def, 0)
	for rows.Next() {
		var d Def
		if err := rows.Scan(&d.ID, &d.EntityType, &d.FieldKey, &d.FieldType, &d.Label, &d.Options, &d.Validation, &d.Required, &d.Searchable, &d.Active); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
