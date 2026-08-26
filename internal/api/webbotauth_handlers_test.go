package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHTTPMessageSignaturesDirectory(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/http-message-signatures-directory", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()

	h.handleHTTPMessageSignaturesDirectory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/http-message-signatures-directory+json") {
		t.Fatalf("Content-Type: %q", ct)
	}
	if rr.Header().Get("Signature") == "" || rr.Header().Get("Signature-Input") == "" {
		t.Fatalf("missing signature headers")
	}
	var doc struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(doc.Keys) < 1 || doc.Keys[0]["kty"] != "OKP" {
		t.Fatalf("keys: %#v", doc.Keys)
	}
}

func TestRouterHTTPMessageSignaturesDirectory(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/http-message-signatures-directory", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%q", rr.Code, rr.Body.String())
	}
}
