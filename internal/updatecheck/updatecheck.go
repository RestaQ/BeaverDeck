package updatecheck

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"time"

	"beaverdeck/internal/config"
	"beaverdeck/internal/kube"
	"beaverdeck/internal/users"
)

type requestPayload struct {
	InstallationID string `json:"installationId"`
	AppVersion     string `json:"appVersion"`
}

type responsePayload struct {
	LatestVersion string `json:"latestVersion"`
}

func Start(ctx context.Context, cfg config.Config, kc *kube.Client, userStore *users.Store) {
	if strings.TrimSpace(cfg.UpdateCheckURL) == "" || strings.TrimSpace(cfg.AppVersion) == "" {
		return
	}
	go loop(ctx, cfg, kc, userStore)
}

func loop(ctx context.Context, cfg config.Config, kc *kube.Client, userStore *users.Store) {
	timer := time.NewTimer(initialDelay(cfg))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			_ = runOnce(ctx, cfg, kc, userStore)
			timer.Reset(nextDelay(cfg))
		}
	}
}

func runOnce(ctx context.Context, cfg config.Config, kc *kube.Client, userStore *users.Store) error {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	installationID, err := installationID(reqCtx, kc, cfg)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(requestPayload{
		InstallationID: installationID,
		AppVersion:     strings.TrimSpace(cfg.AppVersion),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, cfg.UpdateCheckURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var decoded responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil
	}

	err = userStore.SetUpdateCheckStatus(reqCtx, users.UpdateCheckStatus{
		LatestVersion: strings.TrimSpace(decoded.LatestVersion),
	})
	if err != nil {
		return err
	}

	return nil
}

func installationID(ctx context.Context, kc *kube.Client, cfg config.Config) (string, error) {
	namespaceUID, err := kc.NamespaceUID(ctx, cfg.PodNamespace)
	if err != nil {
		return "", err
	}
	serviceAccountUID, err := kc.ServiceAccountUID(ctx, cfg.PodNamespace, cfg.ServiceAccountName)
	if err != nil {
		return "", err
	}
	return namespaceUID + "-" + serviceAccountUID, nil
}

func initialDelay(cfg config.Config) time.Duration {
	jitter := cfg.UpdateCheckJitter
	if jitter <= 0 {
		return time.Minute
	}
	return randomDuration(jitter)
}

func nextDelay(cfg config.Config) time.Duration {
	base := cfg.UpdateCheckEvery
	if base <= 0 {
		base = 24 * time.Hour
	}
	jitter := cfg.UpdateCheckJitter
	if jitter <= 0 {
		return base
	}
	offset := randomDuration(2*jitter) - jitter
	delay := base + offset
	if delay < time.Hour {
		return time.Hour
	}
	return delay
}

func randomDuration(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max.Nanoseconds()+1))
	if err != nil {
		return 0
	}
	return time.Duration(n.Int64())
}
