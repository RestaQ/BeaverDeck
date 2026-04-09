package audit

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Entry struct {
	ID        int64     `json:"id"`
	Time      time.Time `json:"time"`
	Action    string    `json:"action"`
	Namespace string    `json:"namespace"`
	Resource  string    `json:"resource"`
	Name      string    `json:"name"`
	DryRun    bool      `json:"dry_run"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
}

type SuppressedAlert struct {
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "audit.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	stmt := `
CREATE TABLE IF NOT EXISTS audit_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  action TEXT NOT NULL,
  namespace TEXT NOT NULL,
  resource TEXT NOT NULL,
  name TEXT NOT NULL,
  dry_run INTEGER NOT NULL,
  success INTEGER NOT NULL,
  message TEXT NOT NULL
);`
	if _, err := db.Exec(stmt); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate sqlite: %w", err)
	}
	suppressedStmt := `
CREATE TABLE IF NOT EXISTS suppressed_alerts (
  alert_key TEXT PRIMARY KEY,
  created_at TEXT NOT NULL
);`
	if _, err := db.Exec(suppressedStmt); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate suppressed alerts sqlite: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Log(ctx context.Context, e Entry) error {
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_log (ts, action, namespace, resource, name, dry_run, success, message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Time.UTC().Format(time.RFC3339Nano), e.Action, e.Namespace, e.Resource, e.Name,
		boolToInt(e.DryRun), boolToInt(e.Success), e.Message,
	)
	return err
}

func (s *Store) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ts, action, namespace, resource, name, dry_run, success, message
		 FROM audit_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Entry, 0, limit)
	for rows.Next() {
		var (
			e          Entry
			ts         string
			dryRunInt  int
			successInt int
		)
		if err := rows.Scan(&e.ID, &ts, &e.Action, &e.Namespace, &e.Resource, &e.Name, &dryRunInt, &successInt, &e.Message); err != nil {
			return nil, err
		}
		e.DryRun = dryRunInt == 1
		e.Success = successInt == 1
		if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			e.Time = parsed
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (s *Store) ListSuppressedAlerts(ctx context.Context) ([]SuppressedAlert, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT alert_key, created_at FROM suppressed_alerts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SuppressedAlert, 0)
	for rows.Next() {
		var (
			item      SuppressedAlert
			createdAt string
		)
		if err := rows.Scan(&item.Key, &createdAt); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
			item.CreatedAt = parsed
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) SetAlertSuppressed(ctx context.Context, key string, suppressed bool) error {
	if key == "" {
		return fmt.Errorf("alert key is required")
	}
	if suppressed {
		_, err := s.db.ExecContext(
			ctx,
			`INSERT INTO suppressed_alerts (alert_key, created_at) VALUES (?, ?)
			 ON CONFLICT(alert_key) DO UPDATE SET created_at = excluded.created_at`,
			key,
			time.Now().UTC().Format(time.RFC3339Nano),
		)
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM suppressed_alerts WHERE alert_key = ?`, key)
	return err
}
