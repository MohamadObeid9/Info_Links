package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleMCPServerCard(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/mcp/server-card.json", nil)
	rr := httptest.NewRecorder()

	h.handleMCPServerCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	info, ok := doc["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("serverInfo: %#v", doc["serverInfo"])
	}
	if info["name"] == nil || info["version"] == nil {
		t.Fatalf("serverInfo incomplete: %#v", info)
	}
	transport, ok := doc["transport"].(map[string]any)
	if !ok || transport["endpoint"] == nil {
		t.Fatalf("transport: %#v", doc["transport"])
	}
	caps, ok := doc["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities: %#v", doc["capabilities"])
	}
	for _, key := range []string{"tools", "resources", "prompts"} {
		if caps[key] == nil {
			t.Fatalf("capabilities missing %s", key)
		}
	}
}

func TestRouterMCPServerCard(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/mcp/server-card.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
}
