// Package store persists request metrics and app settings in SQLite.
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Request source values stored in the requests.source column.
const (
	SourceManual = "manual"
	SourceProxy  = "proxy"
)

// Store is a SQLite-backed persistence layer for Distilly.
type Store struct {
	db *sql.DB
}

// Request is one logged prompt analysis (manual or proxied).
type Request struct {
	ID              int64   `json:"id"`
	CreatedAt       string  `json:"createdAt"`
	Source          string  `json:"source"`
	Model           string  `json:"model"`
	InputTokens     int     `json:"inputTokens"`
	OptimizedTokens int     `json:"optimizedTokens"`
	SavingsPct      float64 `json:"savingsPct"`
	CostUSD         float64 `json:"costUsd"`
	SavingsUSD      float64 `json:"savingsUsd"`
}

// DashboardStats aggregates savings across all logged requests.
type DashboardStats struct {
	RequestCount int          `json:"requestCount"`
	TokensSaved  int64        `json:"tokensSaved"`
	SavingsUSD   float64      `json:"savingsUsd"`
	ByModel      []ModelStats `json:"byModel"`
}

// ModelStats is per-model aggregate for the dashboard breakdown table.
type ModelStats struct {
	Model        string  `json:"model"`
	RequestCount int     `json:"requestCount"`
	TokensSaved  int64   `json:"tokensSaved"`
	SavingsUSD   float64 `json:"savingsUsd"`
}

// Open opens (or creates) a SQLite database at path and migrates schema.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite needs a single writer; limit pool for file DBs.
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS requests (
  id INTEGER PRIMARY KEY,
  created_at TEXT NOT NULL,
  source TEXT NOT NULL,
  model TEXT,
  input_tokens INTEGER,
  optimized_tokens INTEGER,
  savings_pct REAL,
  cost_usd REAL,
  savings_usd REAL
);

CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

// SetSetting upserts a settings key/value pair.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// GetSetting returns the value for key, or "" if unset.
func (s *Store) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return value, nil
}

// InsertRequest appends a request row. CreatedAt is set to now (UTC RFC3339)
// when empty. Returns the new row id.
func (s *Store) InsertRequest(r Request) (int64, error) {
	createdAt := r.CreatedAt
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	res, err := s.db.Exec(
		`INSERT INTO requests (
			created_at, source, model, input_tokens, optimized_tokens,
			savings_pct, cost_usd, savings_usd
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		createdAt, r.Source, r.Model, r.InputTokens, r.OptimizedTokens,
		r.SavingsPct, r.CostUSD, r.SavingsUSD,
	)
	if err != nil {
		return 0, fmt.Errorf("insert request: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert request id: %w", err)
	}
	return id, nil
}

// GetRecentRequests returns the newest requests, up to limit.
// A non-positive limit defaults to 50.
func (s *Store) GetRecentRequests(limit int) ([]Request, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, created_at, source, model, input_tokens, optimized_tokens,
			savings_pct, cost_usd, savings_usd
		 FROM requests
		 ORDER BY id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent requests: %w", err)
	}
	defer rows.Close()

	out := make([]Request, 0)
	for rows.Next() {
		var r Request
		var model sql.NullString
		if err := rows.Scan(
			&r.ID, &r.CreatedAt, &r.Source, &model,
			&r.InputTokens, &r.OptimizedTokens,
			&r.SavingsPct, &r.CostUSD, &r.SavingsUSD,
		); err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		r.Model = model.String
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recent requests rows: %w", err)
	}
	return out, nil
}

// GetDashboardStats returns totals and per-model breakdown.
func (s *Store) GetDashboardStats() (DashboardStats, error) {
	stats := DashboardStats{ByModel: []ModelStats{}}

	err := s.db.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens - optimized_tokens), 0),
			COALESCE(SUM(savings_usd), 0)
		 FROM requests`,
	).Scan(&stats.RequestCount, &stats.TokensSaved, &stats.SavingsUSD)
	if err != nil {
		return stats, fmt.Errorf("dashboard totals: %w", err)
	}

	rows, err := s.db.Query(
		`SELECT
			COALESCE(model, ''),
			COUNT(*),
			COALESCE(SUM(input_tokens - optimized_tokens), 0),
			COALESCE(SUM(savings_usd), 0)
		 FROM requests
		 GROUP BY model
		 ORDER BY model ASC`,
	)
	if err != nil {
		return stats, fmt.Errorf("dashboard by model: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m ModelStats
		if err := rows.Scan(&m.Model, &m.RequestCount, &m.TokensSaved, &m.SavingsUSD); err != nil {
			return stats, fmt.Errorf("scan model stats: %w", err)
		}
		stats.ByModel = append(stats.ByModel, m)
	}
	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("model stats rows: %w", err)
	}
	return stats, nil
}
