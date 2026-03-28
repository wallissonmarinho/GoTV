package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

func nullTimePG(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// sqlExecutor matches *sql.DB and *sql.Tx for catalog operations.
type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// catalogRepo implements ports.CatalogRepository against a DB or an open transaction.
type catalogRepo struct {
	ex sqlExecutor
	pg bool
}

var _ ports.CatalogRepository = (*catalogRepo)(nil)

func (r *catalogRepo) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	now := time.Now().UTC()
	if r.pg {
		var id int64
		err := r.ex.QueryRowContext(ctx,
			`INSERT INTO m3u_sources (url, label, created_at) VALUES ($1,$2,$3) RETURNING id`,
			url, label, now).Scan(&id)
		if err != nil {
			return nil, err
		}
		return &domain.M3USource{ID: id, URL: url, Label: label, CreatedAt: now}, nil
	}
	res, err := r.ex.ExecContext(ctx,
		`INSERT INTO m3u_sources (url, label, created_at) VALUES (?,?,?)`,
		url, label, now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &domain.M3USource{ID: id, URL: url, Label: label, CreatedAt: now}, nil
}

func (r *catalogRepo) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	rows, err := r.ex.QueryContext(ctx, `SELECT id, url, label, created_at FROM m3u_sources ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.M3USource
	for rows.Next() {
		var s domain.M3USource
		if r.pg {
			if err := rows.Scan(&s.ID, &s.URL, &s.Label, &s.CreatedAt); err != nil {
				return nil, err
			}
		} else {
			var created string
			if err := rows.Scan(&s.ID, &s.URL, &s.Label, &created); err != nil {
				return nil, err
			}
			s.CreatedAt, _ = time.Parse(time.RFC3339, created)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *catalogRepo) DeleteM3USource(ctx context.Context, id int64) error {
	var res sql.Result
	var err error
	if r.pg {
		res, err = r.ex.ExecContext(ctx, `DELETE FROM m3u_sources WHERE id = $1`, id)
	} else {
		res, err = r.ex.ExecContext(ctx, `DELETE FROM m3u_sources WHERE id = ?`, id)
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("m3u source not found")
	}
	return nil
}

func (r *catalogRepo) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	now := time.Now().UTC()
	if r.pg {
		var id int64
		err := r.ex.QueryRowContext(ctx,
			`INSERT INTO epg_sources (url, label, created_at) VALUES ($1,$2,$3) RETURNING id`,
			url, label, now).Scan(&id)
		if err != nil {
			return nil, err
		}
		return &domain.EPGSource{ID: id, URL: url, Label: label, CreatedAt: now}, nil
	}
	res, err := r.ex.ExecContext(ctx,
		`INSERT INTO epg_sources (url, label, created_at) VALUES (?,?,?)`,
		url, label, now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &domain.EPGSource{ID: id, URL: url, Label: label, CreatedAt: now}, nil
}

func (r *catalogRepo) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	rows, err := r.ex.QueryContext(ctx, `SELECT id, url, label, created_at FROM epg_sources ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EPGSource
	for rows.Next() {
		var s domain.EPGSource
		if r.pg {
			if err := rows.Scan(&s.ID, &s.URL, &s.Label, &s.CreatedAt); err != nil {
				return nil, err
			}
		} else {
			var created string
			if err := rows.Scan(&s.ID, &s.URL, &s.Label, &created); err != nil {
				return nil, err
			}
			s.CreatedAt, _ = time.Parse(time.RFC3339, created)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *catalogRepo) DeleteEPGSource(ctx context.Context, id int64) error {
	var res sql.Result
	var err error
	if r.pg {
		res, err = r.ex.ExecContext(ctx, `DELETE FROM epg_sources WHERE id = $1`, id)
	} else {
		res, err = r.ex.ExecContext(ctx, `DELETE FROM epg_sources WHERE id = ?`, id)
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("epg source not found")
	}
	return nil
}

func (r *catalogRepo) SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error {
	m3u := string(snap.M3U)
	epg := string(snap.EPG)
	if r.pg {
		_, err := r.ex.ExecContext(ctx, `
UPDATE merge_snapshot SET
  m3u_text = $1, epg_text = $2, ok = $3, message = $4, last_error = $5,
  channel_count = $6, programme_count = $7, started_at = $8, finished_at = $9
WHERE id = 1`,
			m3u, epg, snap.OK, snap.Message, snap.LastError,
			snap.ChannelCount, snap.ProgrammeCount, nullTimePG(snap.StartedAt), nullTimePG(snap.FinishedAt))
		return err
	}
	_, err := r.ex.ExecContext(ctx, `
UPDATE merge_snapshot SET
  m3u_text = ?, epg_text = ?, ok = ?, message = ?, last_error = ?,
  channel_count = ?, programme_count = ?, started_at = ?, finished_at = ?
WHERE id = 1`,
		m3u, epg, boolToInt(snap.OK), snap.Message, snap.LastError,
		snap.ChannelCount, snap.ProgrammeCount, formatOptionalTime(snap.StartedAt), formatOptionalTime(snap.FinishedAt))
	return err
}

func (r *catalogRepo) LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error) {
	var snap domain.MergeSnapshot
	var m3u, epg, msg, lastErr string
	var ch, prog int
	var ok interface{}
	var started, finished sql.NullString
	if r.pg {
		var startedT, finishedT sql.NullTime
		var okBool bool
		err := r.ex.QueryRowContext(ctx, `
SELECT m3u_text, epg_text, ok, message, last_error, channel_count, programme_count, started_at, finished_at
FROM merge_snapshot WHERE id = 1`).Scan(
			&m3u, &epg, &okBool, &msg, &lastErr, &ch, &prog, &startedT, &finishedT)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return snap, nil
			}
			return snap, err
		}
		snap.M3U = []byte(m3u)
		snap.EPG = []byte(epg)
		snap.Message = msg
		snap.LastError = lastErr
		snap.ChannelCount = ch
		snap.ProgrammeCount = prog
		snap.OK = okBool
		if startedT.Valid {
			snap.StartedAt = startedT.Time
		}
		if finishedT.Valid {
			snap.FinishedAt = finishedT.Time
		}
		return snap, nil
	}
	err := r.ex.QueryRowContext(ctx, `
SELECT m3u_text, epg_text, ok, message, last_error, channel_count, programme_count, started_at, finished_at
FROM merge_snapshot WHERE id = 1`).Scan(
		&m3u, &epg, &ok, &msg, &lastErr, &ch, &prog, &started, &finished)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return snap, nil
		}
		return snap, err
	}
	snap.M3U = []byte(m3u)
	snap.EPG = []byte(epg)
	snap.Message = msg
	snap.LastError = lastErr
	snap.ChannelCount = ch
	snap.ProgrammeCount = prog
	switch v := ok.(type) {
	case int64:
		snap.OK = v != 0
	case bool:
		snap.OK = v
	}
	if started.Valid && started.String != "" {
		snap.StartedAt, _ = time.Parse(time.RFC3339, started.String)
	}
	if finished.Valid && finished.String != "" {
		snap.FinishedAt, _ = time.Parse(time.RFC3339, finished.String)
	}
	return snap, nil
}
