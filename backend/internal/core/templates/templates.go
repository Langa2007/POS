package templates

import (
	"context"
	"errors"
	"html/template"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalid = errors.New("invalid template")

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

type Template struct {
	ID       string  `json:"id"`
	Key      string  `json:"key"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Format   string  `json:"format"`
	Content  string  `json:"content"`
	BranchID *string `json:"branch_id,omitempty"`
	Version  int     `json:"version"`
	Active   bool    `json:"active"`
}

func safeFuncs() template.FuncMap {
	return template.FuncMap{
		"money": func(minor int64, cur string) string { return cur },
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}

func (s *Store) Upsert(ctx context.Context, tx pgx.Tx, tenantID string, t Template) (Template, error) {
	if strings.TrimSpace(t.Key) == "" || strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Content) == "" {
		return Template{}, ErrInvalid
	}
	if _, err := template.New("tpl").Funcs(safeFuncs()).Parse(t.Content); err != nil {
		return Template{}, ErrInvalid
	}
	id := t.ID
	if id == "" {
		nid, err := ids.NewUUIDv7()
		if err != nil {
			return Template{}, err
		}
		id = nid
	}
	ver := t.Version
	if ver < 1 {
		ver = 1
	}
	_, err := tx.Exec(ctx, `INSERT INTO document_templates(id,tenant_id,key,name,type,format,content,branch_id,version,active) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (tenant_id,key,branch_id) DO UPDATE SET name=EXCLUDED.name, type=EXCLUDED.type, format=EXCLUDED.format, content=EXCLUDED.content, version=document_templates.version+1, active=EXCLUDED.active, updated_at=now()`,
		id, tenantID, t.Key, t.Name, t.Type, t.Format, t.Content, t.BranchID, ver, t.Active)
	if err != nil {
		return Template{}, err
	}
	t.ID = id
	t.Version = ver
	return t, nil
}

func (s *Store) Render(ctx context.Context, tenantID, key string, data any) (string, error) {
	return s.render(ctx, s.pool, tenantID, key, data)
}

func (s *Store) RenderTx(ctx context.Context, tx pgx.Tx, tenantID, key string, data any) (string, error) {
	return s.render(ctx, tx, tenantID, key, data)
}

func (s *Store) render(ctx context.Context, queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, tenantID, key string, data any) (string, error) {
	var t Template
	var branchID *string
	if err := queryer.QueryRow(ctx, `SELECT id,key,name,type,format,content,branch_id,version,active FROM document_templates WHERE tenant_id=$1 AND key=$2 AND active=true ORDER BY branch_id NULLS LAST LIMIT 1`, tenantID, key).Scan(&t.ID, &t.Key, &t.Name, &t.Type, &t.Format, &t.Content, &branchID, &t.Version, &t.Active); err != nil {
		return "", err
	}
	tpl, err := template.New("tpl").Funcs(safeFuncs()).Parse(t.Content)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}
