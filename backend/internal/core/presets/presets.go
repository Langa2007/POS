package presets

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrPresetNotFound = errors.New("preset not found")

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

type Preset struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

func (s *Store) List(ctx context.Context) ([]Preset, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,name,data FROM preset_definitions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Preset, 0)
	for rows.Next() {
		var p Preset
		if err := rows.Scan(&p.ID, &p.Name, &p.Data); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (Preset, error) {
	var p Preset
	if err := s.pool.QueryRow(ctx, `SELECT id,name,data FROM preset_definitions WHERE id=$1`, id).Scan(&p.ID, &p.Name, &p.Data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Preset{}, ErrPresetNotFound
		}
		return Preset{}, err
	}
	return p, nil
}

type applyPayload struct {
	Capabilities []string          `json:"capabilities"`
	Roles        []string          `json:"roles"`
	Settings     map[string]any    `json:"settings"`
	CustomFields []json.RawMessage `json:"custom_fields"`
}

func (s *Store) Apply(ctx context.Context, tx pgx.Tx, tenantID, presetID string) error {
	p, err := s.Get(ctx, presetID)
	if err != nil {
		return err
	}
	var payload applyPayload
	if err := json.Unmarshal(p.Data, &payload); err != nil {
		return err
	}
	for _, cap := range payload.Capabilities {
		cap = strings.TrimSpace(cap)
		if cap == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tenant_capabilities(tenant_id,capability_id,enabled) VALUES($1,$2,true) ON CONFLICT (tenant_id,capability_id) DO UPDATE SET enabled=true`, tenantID, cap); err != nil {
			return err
		}
	}
	for _, mod := range p.Modules(presetID) {
		if _, err := tx.Exec(ctx, `INSERT INTO tenant_modules(tenant_id,module_id,enabled) VALUES($1,$2,true) ON CONFLICT (tenant_id,module_id) DO UPDATE SET enabled=true`, tenantID, mod); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO tenant_preset_applied(tenant_id,preset_id,config_snapshot) VALUES($1,$2,$3) ON CONFLICT (tenant_id) DO UPDATE SET preset_id=EXCLUDED.preset_id, config_snapshot=EXCLUDED.config_snapshot, applied_at=now()`, tenantID, presetID, p.Data)
	return err
}

func (p Preset) Modules(id string) []string {
	if strings.Contains(string(p.Data), `"modules"`) {
		var m struct {
			Modules []string `json:"modules"`
		}
		_ = json.Unmarshal(p.Data, &m)
		if len(m.Modules) > 0 {
			out := append([]string(nil), m.Modules...)
			sort.Strings(out)
			return out
		}
	}
	base := []string{"core-checkout"}
	switch id {
	case "pharmacy":
		return append(base, "pharmacy")
	case "restaurant":
		return append(base, "restaurant")
	case "hospitality":
		return append(base, "hospitality")
	}
	return base
}
