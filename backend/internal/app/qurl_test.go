package app

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQURLInlineSVGIsReturnedOnlyAsAnIsolatedImageSource(t *testing.T) {
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"svg":"<svg xmlns=\"http://www.w3.org/2000/svg\"><script>alert(document.domain)</script><rect width=\"10\" height=\"10\"/></svg>"}`))
	}))
	defer providerServer.Close()

	provider := qurlHTTPProvider{baseURL: providerServer.URL, client: providerServer.Client(), allowLocalHTTP: true}
	result, err := provider.GenerateFeedbackQR(context.Background(), "https://heard.example/f/opaque")
	if err != nil {
		t.Fatalf("generate QR: %v", err)
	}
	if result.SVG != "" {
		t.Fatal("provider markup must not remain available for DOM injection")
	}
	const prefix = "data:image/svg+xml;base64,"
	if !strings.HasPrefix(result.AssetURL, prefix) {
		t.Fatalf("asset URL %q is not an isolated SVG image", result.AssetURL)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(result.AssetURL, prefix))
	if err != nil || !strings.Contains(string(decoded), "<svg") {
		t.Fatalf("isolated SVG payload is invalid: %v", err)
	}
}

func TestQURLRejectsExecutableAssetURLScheme(t *testing.T) {
	if _, err := normalizeQURLResult(QURLResult{AssetURL: "javascript:alert(1)"}, false); err == nil {
		t.Fatal("expected executable asset URL to be rejected")
	}
}

func TestQURLRejectsPublicHTTPAssetURL(t *testing.T) {
	if _, err := normalizeQURLResult(QURLResult{AssetURL: "http://cdn.example/qr.svg"}, true); err == nil {
		t.Fatal("expected public cleartext asset URL to be rejected even in local mode")
	}
	if _, err := normalizeQURLResult(QURLResult{AssetURL: "http://127.0.0.1:8082/qr.svg"}, true); err != nil {
		t.Fatalf("explicit local provider asset should be accepted in local mode: %v", err)
	}
}

func TestStoredProviderMarkupIsNeverReturnedRaw(t *testing.T) {
	link := FeedbackLink{QRSVG: `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`}
	isolateStoredQURL(&link, false)
	if link.QRSVG != "" || !strings.HasPrefix(link.QRAssetURL, "data:image/svg+xml;base64,") {
		t.Fatalf("stored SVG was not isolated: %#v", link)
	}

	link = FeedbackLink{QRAssetURL: "http://cdn.example/legacy.svg"}
	isolateStoredQURL(&link, true)
	if link.QRAssetURL != "" || link.QRSVG != "" {
		t.Fatalf("unsafe legacy asset should be removed: %#v", link)
	}
}
