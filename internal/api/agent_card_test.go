package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAgentCard(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	rr := httptest.NewRecorder()

	h.handleAgentCard(rr, req)

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
	for _, key := range []string{"name", "version", "description", "supportedInterfaces", "capabilities", "skills"} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
	if doc["name"] != "Info Links" {
		t.Fatalf("name: %#v", doc["name"])
	}
	ifaces, ok := doc["supportedInterfaces"].([]any)
	if !ok || len(ifaces) == 0 {
		t.Fatalf("supportedInterfaces: %#v", doc["supportedInterfaces"])
	}
	iface, _ := ifaces[0].(map[string]any)
	if iface["url"] == nil || iface["protocolBinding"] == nil {
		t.Fatalf("interface incomplete: %#v", iface)
	}
	skills, ok := doc["skills"].([]any)
	if !ok || len(skills) == 0 {
		t.Fatalf("skills: %#v", doc["skills"])
	}
	skill, _ := skills[0].(map[string]any)
	for _, key := range []string{"id", "name", "description"} {
		if skill[key] == nil {
			t.Fatalf("skill missing %s: %#v", key, skill)
		}
	}
}

func TestRouterAgentCard(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
}
