package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	appStateInitializedKey    = "initialized"
	appStateBootstrapTokenKey = "bootstrap_token"
)

func (s *Store) EnsureDefaults(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO roles (name, mode, permissions, created_at) VALUES ('admin', 'admin', '{}', ?)`, now); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO roles (name, mode, permissions, created_at) VALUES ('viewer', 'viewer', '{}', ?)`, now); err != nil {
		return err
	}
	return nil
}

func (s *Store) PrepareBootstrap(ctx context.Context) (BootstrapStatus, error) {
	if err := s.EnsureDefaults(ctx); err != nil {
		return BootstrapStatus{}, err
	}

	status, err := s.GetBootstrapStatus(ctx)
	if err != nil {
		return BootstrapStatus{}, err
	}
	if status.Initialized {
		return status, nil
	}

	legacyAdmin, err := s.hasLegacyLocalAdmin(ctx)
	if err != nil {
		return BootstrapStatus{}, err
	}
	if legacyAdmin {
		if err := s.setAppState(ctx, appStateInitializedKey, "true"); err != nil {
			return BootstrapStatus{}, err
		}
		if err := s.setAppState(ctx, appStateBootstrapTokenKey, ""); err != nil {
			return BootstrapStatus{}, err
		}
		return BootstrapStatus{Initialized: true}, nil
	}

	token, err := randomToken()
	if err != nil {
		return BootstrapStatus{}, err
	}
	if err := s.setAppState(ctx, appStateInitializedKey, "false"); err != nil {
		return BootstrapStatus{}, err
	}
	if err := s.setAppState(ctx, appStateBootstrapTokenKey, token); err != nil {
		return BootstrapStatus{}, err
	}
	return BootstrapStatus{Initialized: false, Token: token}, nil
}

func (s *Store) GetBootstrapStatus(ctx context.Context) (BootstrapStatus, error) {
	initializedValue, err := s.getAppState(ctx, appStateInitializedKey)
	if err != nil {
		return BootstrapStatus{}, err
	}
	tokenValue, err := s.getAppState(ctx, appStateBootstrapTokenKey)
	if err != nil {
		return BootstrapStatus{}, err
	}
	return BootstrapStatus{
		Initialized: strings.EqualFold(strings.TrimSpace(initializedValue), "true"),
		Token:       strings.TrimSpace(tokenValue),
	}, nil
}

func (s *Store) CompleteBootstrap(ctx context.Context, bootstrapToken, adminPassword string) error {
	bootstrapToken = strings.TrimSpace(bootstrapToken)
	adminPassword = strings.TrimSpace(adminPassword)
	if bootstrapToken == "" || adminPassword == "" {
		return fmt.Errorf("bootstrap token and admin password are required")
	}
	passwordHash, err := hashLocalPassword(adminPassword)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.ensureDefaultsTx(ctx, tx); err != nil {
		return err
	}

	initializedValue, err := getAppStateTx(ctx, tx, appStateInitializedKey)
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(initializedValue), "true") {
		return fmt.Errorf("application is already initialized")
	}

	expectedToken, err := getAppStateTx(ctx, tx, appStateBootstrapTokenKey)
	if err != nil {
		return err
	}
	if strings.TrimSpace(expectedToken) == "" || bootstrapToken != strings.TrimSpace(expectedToken) {
		return fmt.Errorf("invalid bootstrap token")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO users (username, token, role, auth_source, google_subject, session_version, created_at)
		 VALUES ('admin', ?, ?, 'local', '', 1, ?)
		 ON CONFLICT(username) DO UPDATE SET
		   token = excluded.token,
		   role = excluded.role,
		   auth_source = 'local',
		   google_subject = '',
		   session_version = CASE WHEN users.token = excluded.token THEN users.session_version ELSE users.session_version + 1 END`,
		passwordHash, string(RoleAdmin), now,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE username = 'admin'`); err != nil {
		return err
	}
	if err := setAppStateTx(ctx, tx, appStateInitializedKey, "true"); err != nil {
		return err
	}
	if err := setAppStateTx(ctx, tx, appStateBootstrapTokenKey, ""); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ensureDefaultsTx(ctx context.Context, tx *sql.Tx) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO roles (name, mode, permissions, created_at) VALUES ('admin', 'admin', '{}', ?)`, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO roles (name, mode, permissions, created_at) VALUES ('viewer', 'viewer', '{}', ?)`, now); err != nil {
		return err
	}
	return nil
}

func (s *Store) hasLegacyLocalAdmin(ctx context.Context) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT CASE WHEN EXISTS(
			SELECT 1 FROM users WHERE username = 'admin' AND auth_source = 'local' AND TRIM(token) <> ''
		) THEN 1 ELSE 0 END`,
	).Scan(&exists)
	return exists == 1, err
}

func (s *Store) getAppState(ctx context.Context, key string) (string, error) {
	return getAppStateQuerier(ctx, s.db, key)
}

func getAppStateTx(ctx context.Context, tx *sql.Tx, key string) (string, error) {
	return getAppStateQuerier(ctx, tx, key)
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func getAppStateQuerier(ctx context.Context, q queryRower, key string) (string, error) {
	var value string
	err := q.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) setAppState(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO app_state (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func setAppStateTx(ctx context.Context, tx *sql.Tx, key, value string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO app_state (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}
