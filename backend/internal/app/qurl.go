package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type QURLResult struct {
	AssetURL string `json:"asset_url"`
	SVG      string `json:"svg"`
}

type QURLProvider interface {
	GenerateFeedbackQR(ctx context.Context, destinationURL string) (QURLResult, error)
}

type qurlHTTPProvider struct {
	baseURL string
	client  *http.Client
}

type qurlDisabledProvider struct{}

func NewQURLProvider(cfg Config) QURLProvider {
	if strings.TrimSpace(cfg.QURLBaseURL) == "" {
		return qurlDisabledProvider{}
	}
	return qurlHTTPProvider{
		baseURL: strings.TrimRight(cfg.QURLBaseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (qurlDisabledProvider) GenerateFeedbackQR(_ context.Context, _ string) (QURLResult, error) {
	return QURLResult{}, errors.New("qurl is not configured; set QURL_BASE_URL to enable QR asset generation")
}

func (p qurlHTTPProvider) GenerateFeedbackQR(ctx context.Context, destinationURL string) (QURLResult, error) {
	payload, err := json.Marshal(map[string]any{
		"destination_url": destinationURL,
		"format":          "svg",
		"style": map[string]any{
			"preset": "print-flyer",
		},
	})
	if err != nil {
		return QURLResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/qr-codes", bytes.NewReader(payload))
	if err != nil {
		return QURLResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return QURLResult{}, fmt.Errorf("call qurl: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return QURLResult{}, fmt.Errorf("qurl returned status %d", resp.StatusCode)
	}

	var result QURLResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return QURLResult{}, fmt.Errorf("decode qurl response: %w", err)
	}
	if result.AssetURL == "" && result.SVG == "" {
		return QURLResult{}, errors.New("qurl response did not include an asset URL or SVG")
	}
	return result, nil
}
