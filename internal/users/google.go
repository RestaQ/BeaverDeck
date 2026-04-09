package users

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (s *Store) GetGoogleConfig(ctx context.Context) (GoogleConfig, error) {
	var cfg GoogleConfig
	err := s.db.QueryRowContext(ctx,
		`SELECT client_id, client_secret, hosted_domain, service_account_json, delegated_admin_email, updated_at
		 FROM google_config WHERE singleton = 1`,
	).Scan(&cfg.ClientID, &cfg.ClientSecret, &cfg.HostedDomain, &cfg.ServiceAccountJSON, &cfg.DelegatedAdminEmail, &cfg.UpdatedAt)
	return cfg, err
}

func (s *Store) UpdateGoogleConfig(ctx context.Context, cfg GoogleConfig) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE google_config
		 SET client_id = ?, client_secret = ?, hosted_domain = ?, service_account_json = ?, delegated_admin_email = ?, updated_at = ?
		 WHERE singleton = 1`,
		strings.TrimSpace(cfg.ClientID),
		strings.TrimSpace(cfg.ClientSecret),
		strings.TrimSpace(strings.ToLower(cfg.HostedDomain)),
		strings.TrimSpace(cfg.ServiceAccountJSON),
		strings.TrimSpace(strings.ToLower(cfg.DelegatedAdminEmail)),
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) ListGoogleGroupRoles(ctx context.Context) ([]GoogleGroupRole, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT group_email, role, created_at
		 FROM google_group_roles
		 ORDER BY group_email ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []GoogleGroupRole
	for rows.Next() {
		var (
			item      GoogleGroupRole
			createdAt string
		)
		if err := rows.Scan(&item.GroupEmail, &item.Role, &createdAt); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
			item.CreatedAt = parsed
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) UpsertGoogleGroupRole(ctx context.Context, groupEmail string, role Role) error {
	groupEmail = strings.TrimSpace(strings.ToLower(groupEmail))
	if groupEmail == "" {
		return fmt.Errorf("google group email is required")
	}
	if !s.roleExists(ctx, string(role)) {
		return fmt.Errorf("role does not exist: %s", role)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO google_group_roles (group_email, role, created_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(group_email) DO UPDATE SET role = excluded.role`,
		groupEmail, string(role), time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) DeleteGoogleGroupRole(ctx context.Context, groupEmail string) error {
	groupEmail = strings.TrimSpace(strings.ToLower(groupEmail))
	if groupEmail == "" {
		return fmt.Errorf("google group email is required")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM google_group_roles WHERE group_email = ?`, groupEmail)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ResetGoogleAuth(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`UPDATE google_config
		 SET client_id = '', client_secret = '', hosted_domain = '', service_account_json = '', delegated_admin_email = '', directory_token = '', updated_at = ?
		 WHERE singleton = 1`,
		time.Now().UTC().Format(time.RFC3339Nano),
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM google_group_roles`); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ResolveGoogleRole(ctx context.Context, groups []string) (Role, string, error) {
	normalized := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(strings.ToLower(group))
		if group != "" {
			normalized[group] = struct{}{}
		}
	}
	if len(normalized) == 0 {
		return "", "", fmt.Errorf("google user is not a member of any mapped group")
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT m.group_email, m.role,
		        COALESCE(r.mode, CASE WHEN m.role = 'admin' THEN 'admin' ELSE 'viewer' END) AS mode
		 FROM google_group_roles m
		 LEFT JOIN roles r ON r.name = m.role
		 ORDER BY CASE WHEN COALESCE(r.mode, CASE WHEN m.role = 'admin' THEN 'admin' ELSE 'viewer' END) = 'admin' THEN 0 ELSE 1 END,
		          m.role ASC,
		          m.group_email ASC`,
	)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			groupEmail string
			role       Role
			mode       string
		)
		if err := rows.Scan(&groupEmail, &role, &mode); err != nil {
			return "", "", err
		}
		if _, ok := normalized[strings.ToLower(groupEmail)]; !ok {
			continue
		}
		return role, groupEmail, nil
	}
	if err := rows.Err(); err != nil {
		return "", "", err
	}
	return "", "", fmt.Errorf("google user is not a member of any mapped group")
}
