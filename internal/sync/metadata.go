package sync

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/parth/os-agent/internal/config"
)

// MetadataSyncer handles syncing schema.json to the OutSystems REST endpoint.
type MetadataSyncer struct {
	client *http.Client
	cfg    *config.Config
}

// NewMetadataSyncer creates a syncer with an optimized HTTP client for connection reuse.
func NewMetadataSyncer(cfg *config.Config) *MetadataSyncer {
	return &MetadataSyncer{
		cfg: cfg,
		client: &http.Client{
			Timeout: 300 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

// Sync reads schema.json and POSTs it to the configured OutSystems endpoint.
func (m *MetadataSyncer) Sync(schemaPath string) error {
	if m.cfg.BaseURL == "" || m.cfg.AuthToken == "" {
		return fmt.Errorf("OS_URL or OS_AUTH_TOKEN not configured — skipping metadata sync")
	}

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	url := fmt.Sprintf("%s%s", m.cfg.BaseURL, m.cfg.SyncEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "os-agent/0.1")

	// Set auth header based on configured mode.
	switch strings.ToLower(m.cfg.AuthMode) {
	case "basic":
		// OS_AUTH_TOKEN should be "username:password" for basic auth.
		encoded := base64.StdEncoding.EncodeToString([]byte(m.cfg.AuthToken))
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	default:
		// Bearer token (e.g., LifeTime Service Account token).
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.cfg.AuthToken))
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("metadata sync request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("  ✅ Schema synced → %s [%d]\n", url, resp.StatusCode)
		return nil
	}

	// Read up to 1KB of the error response body for debugging.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("metadata sync failed [%d]: %s", resp.StatusCode, string(body))
}
