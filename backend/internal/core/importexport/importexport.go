package importexport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/Langa2007/POS/backend/internal/kernel/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalid = errors.New("invalid import")

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

type Job struct {
	ID         string          `json:"id"`
	EntityType string          `json:"entity_type"`
	Status     string          `json:"status"`
	Mapping    json.RawMessage `json:"mapping"`
	Filename   string          `json:"filename"`
	Errors     json.RawMessage `json:"errors"`
	Processed  int             `json:"processed"`
}

func (s *Store) Create(ctx context.Context, tx pgx.Tx, tenantID, entityType, filename string, mapping json.RawMessage) (Job, error) {
	if strings.TrimSpace(entityType) == "" {
		return Job{}, ErrInvalid
	}
	if mapping == nil {
		mapping = json.RawMessage("{}")
	}
	id, err := ids.NewUUIDv7()
	if err != nil {
		return Job{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO import_jobs(id,tenant_id,entity_type,status,mapping,filename,errors,processed) VALUES($1,$2,$3,'pending',$4,$5,'[]'::jsonb,0)`,
		id, tenantID, entityType, mapping, filename)
	if err != nil {
		return Job{}, err
	}
	return Job{ID: id, EntityType: entityType, Status: "pending", Mapping: mapping, Filename: filename, Errors: json.RawMessage("[]"), Processed: 0}, nil
}

func ParseCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrInvalid
	}
	return rows, nil
}

func (s *Store) DryRun(ctx context.Context, tx pgx.Tx, jobID string, rows [][]string) ([]string, error) {
	errs := make([]string, 0)
	if len(rows) < 2 {
		errs = append(errs, "no_data_rows")
	}
	_, err := tx.Exec(ctx, `UPDATE import_jobs SET status='dry_run', errors=$2 WHERE id=$1`, jobID, toJSON(errs))
	return errs, err
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
