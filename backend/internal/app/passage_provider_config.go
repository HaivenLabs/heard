package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ConfigurePassageGoogle(ctx context.Context, cfg Config, client *http.Client) error {
	if strings.TrimSpace(cfg.GoogleClientID) == "" && strings.TrimSpace(cfg.GoogleClientSecret) == "" {
		return nil
	}
	payload, err := json.Marshal(map[string]string{
		"provider_client_id":     cfg.GoogleClientID,
		"provider_client_secret": cfg.GoogleClientSecret,
		"redirect_uri":           cfg.GoogleRedirectURL,
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(cfg.PassageBaseURL, "/") + "/api/v1/internal/applications/" + url.PathEscape(cfg.PassageClientID) + "/identity-providers/google"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.PassageProviderConfigurationToken)
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("configure Passage Google provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return fmt.Errorf("configure Passage Google provider: HTTP %d", resp.StatusCode)
	}
	return nil
}
