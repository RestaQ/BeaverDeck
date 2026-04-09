package users

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "users.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}

	migrations := []struct {
		stmt string
		msg  string
	}{
		{
			stmt: `
CREATE TABLE IF NOT EXISTS app_state (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT ''
);`,
			msg: "migrate app_state sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS users (
  username TEXT PRIMARY KEY,
  token TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL,
  created_at TEXT NOT NULL
);`,
			msg: "migrate sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS roles (
  name TEXT PRIMARY KEY,
  mode TEXT NOT NULL,
  permissions TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);`,
			msg: "migrate roles sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  auth_source TEXT NOT NULL,
  session_version INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(username) REFERENCES users(username) ON DELETE CASCADE
);`,
			msg: "migrate sessions sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS google_config (
  singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
  client_id TEXT NOT NULL DEFAULT '',
  client_secret TEXT NOT NULL DEFAULT '',
  hosted_domain TEXT NOT NULL DEFAULT '',
  service_account_json TEXT NOT NULL DEFAULT '',
  delegated_admin_email TEXT NOT NULL DEFAULT '',
  directory_token TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT ''
);`,
			msg: "migrate google_config sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS google_group_roles (
  group_email TEXT PRIMARY KEY,
  role TEXT NOT NULL,
  created_at TEXT NOT NULL
);`,
			msg: "migrate google_group_roles sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS oidc_config (
  singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
  provider_name TEXT NOT NULL DEFAULT '',
  issuer_url TEXT NOT NULL DEFAULT '',
  client_id TEXT NOT NULL DEFAULT '',
  client_secret TEXT NOT NULL DEFAULT '',
  scopes TEXT NOT NULL DEFAULT 'openid email profile groups',
  hosted_domain TEXT NOT NULL DEFAULT '',
  email_claim TEXT NOT NULL DEFAULT 'email',
  groups_claim TEXT NOT NULL DEFAULT 'groups',
  updated_at TEXT NOT NULL DEFAULT ''
);`,
			msg: "migrate oidc_config sqlite",
		},
		{
			stmt: `
CREATE TABLE IF NOT EXISTS oidc_group_roles (
  group_name TEXT PRIMARY KEY,
  role TEXT NOT NULL,
  created_at TEXT NOT NULL
);`,
			msg: "migrate oidc_group_roles sqlite",
		},
	}
	for _, migration := range migrations {
		if _, err := db.Exec(migration.stmt); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%s: %w", migration.msg, err)
		}
	}

	addColumn := func(stmt, msg string) error {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			_ = db.Close()
			return fmt.Errorf("%s: %w", msg, err)
		}
		return nil
	}

	if err := addColumn(`ALTER TABLE roles ADD COLUMN permissions TEXT NOT NULL DEFAULT '{}'`, "migrate roles permissions column"); err != nil {
		return nil, err
	}
	if err := addColumn(`ALTER TABLE users ADD COLUMN auth_source TEXT NOT NULL DEFAULT 'local'`, "migrate users auth_source column"); err != nil {
		return nil, err
	}
	if err := addColumn(`ALTER TABLE users ADD COLUMN google_subject TEXT NOT NULL DEFAULT ''`, "migrate users google_subject column"); err != nil {
		return nil, err
	}
	if err := addColumn(`ALTER TABLE users ADD COLUMN session_version INTEGER NOT NULL DEFAULT 1`, "migrate users session_version column"); err != nil {
		return nil, err
	}
	if err := addColumn(`ALTER TABLE google_config ADD COLUMN service_account_json TEXT NOT NULL DEFAULT ''`, "migrate google_config service_account_json column"); err != nil {
		return nil, err
	}
	if err := addColumn(`ALTER TABLE google_config ADD COLUMN delegated_admin_email TEXT NOT NULL DEFAULT ''`, "migrate google_config delegated_admin_email column"); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`INSERT OR IGNORE INTO google_config (singleton, client_id, client_secret, hosted_domain, service_account_json, delegated_admin_email, directory_token, updated_at) VALUES (1, '', '', '', '', '', '', '')`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seed google_config sqlite: %w", err)
	}
	if _, err := db.Exec(`INSERT OR IGNORE INTO oidc_config (singleton, provider_name, issuer_url, client_id, client_secret, scopes, hosted_domain, email_claim, groups_claim, updated_at) VALUES (1, '', '', '', '', 'openid email profile groups', '', 'email', 'groups', '')`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("seed oidc_config sqlite: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
