package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	baseURL        string
	client         *http.Client
	allowLocalHTTP bool
}

type qurlDisabledProvider struct{}

const (
	maxQURLResponseBytes = 512 * 1024
	maxQURLSVGBytes      = 256 * 1024
)

func NewQURLProvider(cfg Config) QURLProvider {
	if strings.TrimSpace(cfg.QURLBaseURL) == "" {
		return qurlDisabledProvider{}
	}
	baseURL := strings.TrimRight(cfg.QURLBaseURL, "/")
	trusted, _ := url.Parse(baseURL)
	return qurlHTTPProvider{
		baseURL: baseURL,
		client: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || trusted == nil || !strings.EqualFold(req.URL.Host, trusted.Host) || req.URL.Scheme != trusted.Scheme {
				return errors.New("untrusted qurl redirect")
			}
			return nil
		}},
		allowLocalHTTP: cfg.IsLocalRuntime(),
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

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxQURLResponseBytes+1))
	if err != nil {
		return QURLResult{}, fmt.Errorf("read qurl response: %w", err)
	}
	if len(raw) > maxQURLResponseBytes {
		return QURLResult{}, errors.New("qurl response exceeds the maximum size")
	}
	var result QURLResult
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&result); err != nil {
		return QURLResult{}, fmt.Errorf("decode qurl response: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return QURLResult{}, errors.New("decode qurl response: response must contain one JSON object")
	}
	return normalizeQURLResult(result, p.allowLocalHTTP)
}

// normalizeQURLResult ensures provider markup can only be consumed as an image
// document. SVG returned by qurl is never exposed for insertion into Heard's DOM.
func normalizeQURLResult(result QURLResult, allowLocalHTTP bool) (QURLResult, error) {
	result.AssetURL = strings.TrimSpace(result.AssetURL)
	if result.AssetURL != "" {
		if len(result.AssetURL) > 4096 {
			return QURLResult{}, errors.New("qurl asset URL is too long")
		}
		parsed, err := url.Parse(result.AssetURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return QURLResult{}, errors.New("qurl asset URL must be an absolute HTTPS URL")
		}
		if parsed.Scheme == "http" && (!allowLocalHTTP || !isLoopbackURL(result.AssetURL)) {
			return QURLResult{}, errors.New("qurl asset URL must use HTTPS unless it is an explicit local provider asset")
		}
		return QURLResult{AssetURL: result.AssetURL}, nil
	}

	svg := strings.TrimSpace(result.SVG)
	if svg == "" {
		return QURLResult{}, errors.New("qurl response did not include an asset URL or SVG")
	}
	if len(svg) > maxQURLSVGBytes {
		return QURLResult{}, errors.New("qurl SVG exceeds the maximum asset size")
	}
	lower := strings.ToLower(svg)
	if !strings.HasPrefix(lower, "<svg") || !strings.HasSuffix(lower, "</svg>") {
		return QURLResult{}, errors.New("qurl SVG payload is malformed")
	}
	return QURLResult{AssetURL: "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))}, nil
}

func isolateStoredQURL(link *FeedbackLink, allowLocalHTTP bool) {
	stored := QURLResult{AssetURL: link.QRAssetURL, SVG: link.QRSVG}
	link.QRAssetURL = ""
	link.QRSVG = ""
	normalized, err := normalizeQURLResult(stored, allowLocalHTTP)
	if err == nil {
		link.QRAssetURL = normalized.AssetURL
	}
}
