package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAgentsIndex(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agents-index.json", nil)
	rr := httptest.NewRecorder()

	h.handleAgentsIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" || ct[:16] != "application/json" {
		t.Fatalf("Content-Type: got %q", ct)
	}

	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if doc["origin"] == nil || doc["version"] == nil {
		t.Fatalf("missing origin/version: %#v", doc)
	}
	agents, ok := doc["agents"].(map[string]any)
	if !ok {
		t.Fatalf("agents: %#v", doc["agents"])
	}
	entry, ok := agents["info-links"].(map[string]any)
	if !ok {
		t.Fatalf("info-links: %#v", agents)
	}
	loc, ok := entry["location"].(map[string]any)
	if !ok {
		t.Fatalf("location: %#v", entry["location"])
	}
	wk, ok := loc["wellKnown"].(map[string]any)
	if !ok || wk["a2a"] == nil || wk["mcp"] == nil {
		t.Fatalf("wellKnown: %#v", loc["wellKnown"])
	}
	cap, ok := entry["capability"].(map[string]any)
	if !ok {
		t.Fatalf("capability: %#v", entry["capability"])
	}
	protos, ok := cap["protocols"].([]any)
	if !ok || len(protos) < 2 {
		t.Fatalf("protocols: %#v", cap["protocols"])
	}
}

func TestRouterAgentsIndex(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agents-index.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
}
