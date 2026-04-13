package users

import (
	"context"
	"strings"
)

const (
	appStateUpdateLatestVersionKey = "update_latest_version"
)

type UpdateCheckStatus struct {
	LatestVersion string `json:"latestVersion"`
}

func (s *Store) GetUpdateCheckStatus(ctx context.Context) (UpdateCheckStatus, error) {
	latestVersion, err := s.getAppState(ctx, appStateUpdateLatestVersionKey)
	if err != nil {
		return UpdateCheckStatus{}, err
	}

	return UpdateCheckStatus{
		LatestVersion: strings.TrimSpace(latestVersion),
	}, nil
}

func (s *Store) SetUpdateCheckStatus(ctx context.Context, status UpdateCheckStatus) error {
	return s.setAppState(ctx, appStateUpdateLatestVersionKey, strings.TrimSpace(status.LatestVersion))
}
