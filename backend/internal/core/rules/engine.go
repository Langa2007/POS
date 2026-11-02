package rules

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalid = errors.New("invalid rule")

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

type Rule struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Enabled   bool            `json:"enabled"`
	Priority  int             `json:"priority"`
	Condition json.RawMessage `json:"condition"`
	Actions   json.RawMessage `json:"actions"`
	ValidFrom *time.Time      `json:"valid_from,omitempty"`
	ValidTo   *time.Time      `json:"valid_to,omitempty"`
}

type EvalRequest struct {
	Facts map[string]any `json:"facts"`
}

type Action struct {
	Type  string         `json:"type"`
	Value map[string]any `json:"value"`
}

func (s *Store) Upsert(ctx context.Context, tx pgx.Tx, tenantID string, r Rule) (Rule, error) {
	if strings.TrimSpace(r.Name) == "" || len(r.Condition) == 0 || len(r.Actions) == 0 {
		return Rule{}, ErrInvalid
	}
	id := r.ID
	if id == "" {
		nid, err := ids.NewUUIDv7()
		if err != nil {
			return Rule{}, err
		}
		id = nid
	}
	_, err := tx.Exec(ctx, `INSERT INTO rules(id,tenant_id,name,enabled,priority,condition,actions,valid_from,valid_to) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, enabled=EXCLUDED.enabled, priority=EXCLUDED.priority, condition=EXCLUDED.condition, actions=EXCLUDED.actions, valid_from=EXCLUDED.valid_from, valid_to=EXCLUDED.valid_to, updated_at=now()`,
		id, tenantID, r.Name, r.Enabled, r.Priority, r.Condition, r.Actions, r.ValidFrom, r.ValidTo)
	if err != nil {
		return Rule{}, err
	}
	r.ID = id
	return r, nil
}

func (s *Store) List(ctx context.Context, tenantID string, enabledOnly bool) ([]Rule, error) {
	return s.list(ctx, s.pool, tenantID, enabledOnly)
}

func (s *Store) ListTx(ctx context.Context, tx pgx.Tx, tenantID string, enabledOnly bool) ([]Rule, error) {
	return s.list(ctx, tx, tenantID, enabledOnly)
}

func (s *Store) list(ctx context.Context, queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, tenantID string, enabledOnly bool) ([]Rule, error) {
	q := `SELECT id,name,enabled,priority,condition,actions,valid_from,valid_to FROM rules WHERE tenant_id=$1`
	if enabledOnly {
		q += ` AND enabled=true`
	}
	q += ` ORDER BY priority DESC, created_at ASC`
	rows, err := s.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Rule, 0)
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.ID, &r.Name, &r.Enabled, &r.Priority, &r.Condition, &r.Actions, &r.ValidFrom, &r.ValidTo); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func getFact(facts map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = facts
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func cmp(a, b any, op string) bool {
	switch op {
	case "eq":
		return fmtAny(a) == fmtAny(b)
	case "ne":
		return fmtAny(a) != fmtAny(b)
	case "exists":
		return a != nil
	}
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if !aok || !bok {
		if op == "eq" {
			return fmtAny(a) == fmtAny(b)
		}
		return false
	}
	switch op {
	case "gt":
		return af > bf
	case "gte":
		return af >= bf
	case "lt":
		return af < bf
	case "lte":
		return af <= bf
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func fmtAny(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

type cond struct {
	All   []cond `json:"all"`
	Any   []cond `json:"any"`
	Not   *cond  `json:"not"`
	Fact  string `json:"fact"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

func evalCond(c cond, facts map[string]any) bool {
	if len(c.All) > 0 {
		for _, x := range c.All {
			if !evalCond(x, facts) {
				return false
			}
		}
		return true
	}
	if len(c.Any) > 0 {
		for _, x := range c.Any {
			if evalCond(x, facts) {
				return true
			}
		}
		return false
	}
	if c.Not != nil {
		return !evalCond(*c.Not, facts)
	}
	if c.Fact == "" {
		return true
	}
	val, _ := getFact(facts, c.Fact)
	return cmp(val, c.Value, c.Op)
}

func (s *Store) Simulate(ctx context.Context, tenantID string, req EvalRequest, now time.Time) ([]map[string]any, error) {
	return s.simulate(ctx, s.pool, tenantID, req, now)
}

func (s *Store) SimulateTx(ctx context.Context, tx pgx.Tx, tenantID string, req EvalRequest, now time.Time) ([]map[string]any, error) {
	return s.simulate(ctx, tx, tenantID, req, now)
}

func (s *Store) simulate(ctx context.Context, queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, tenantID string, req EvalRequest, now time.Time) ([]map[string]any, error) {
	rules, err := s.list(ctx, queryer, tenantID, true)
	if err != nil {
		return nil, err
	}
	matched := make([]map[string]any, 0)
	for _, r := range rules {
		if r.ValidFrom != nil && now.Before(*r.ValidFrom) {
			continue
		}
		if r.ValidTo != nil && now.After(*r.ValidTo) {
			continue
		}
		var c cond
		if err := json.Unmarshal(r.Condition, &c); err != nil {
			continue
		}
		if !evalCond(c, req.Facts) {
			continue
		}
		var acts []map[string]any
		_ = json.Unmarshal(r.Actions, &acts)
		matched = append(matched, map[string]any{"rule_id": r.ID, "rule_name": r.Name, "priority": r.Priority, "actions": acts})
	}
	return matched, nil
}

func ApplyDiscounts(totalMinor int64, matched []map[string]any) int64 {
	res := totalMinor
	for _, m := range matched {
		acts, _ := m["actions"].([]any)
		for _, a := range acts {
			am, ok := a.(map[string]any)
			if !ok {
				continue
			}
			t, _ := am["type"].(string)
			if t == "set_discount_percent" {
				if v, ok := am["value"].(map[string]any); ok {
					if p, ok := toFloat(v["percent"]); ok && p > 0 && p <= 100 {
						res = int64(math.Round(float64(res) * (1.0 - p/100.0)))
					}
				}
			} else if t == "set_discount_fixed" {
				if v, ok := am["value"].(map[string]any); ok {
					if d, ok := toFloat(v["amount_minor"]); ok && d > 0 {
						res -= int64(d)
					}
				}
			}
		}
	}
	if res < 0 {
		res = 0
	}
	return res
}
