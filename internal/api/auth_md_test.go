package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAuthMD(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/auth.md", nil)
	rr := httptest.NewRecorder()

	h.handleAuthMD(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "text/markdown") {
		t.Fatalf("Content-Type: got %q", rr.Header().Get("Content-Type"))
	}
	body := rr.Body.String()
	if !strings.HasPrefix(body, "# auth.md\n") {
		t.Fatalf("expected H1 auth.md, got: %q", body[:min(40, len(body))])
	}
	if !strings.Contains(body, "/api/users/guest") || !strings.Contains(body, "anonymous") {
		t.Fatalf("missing registration guidance: %s", body[:min(200, len(body))])
	}
}

func TestHandleOAuthProtectedResource(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	rr := httptest.NewRecorder()

	h.handleOAuthProtectedResource(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if doc["resource"] != "https://example.com/" {
		t.Fatalf("resource: got %#v", doc["resource"])
	}
	servers, _ := doc["authorization_servers"].([]any)
	if len(servers) == 0 || servers[0] != "https://example.com" {
		t.Fatalf("authorization_servers: %#v", doc["authorization_servers"])
	}
	methods, _ := doc["bearer_methods_supported"].([]any)
	if len(methods) == 0 || methods[0] != "header" {
		t.Fatalf("bearer_methods_supported: %#v", doc["bearer_methods_supported"])
	}
	if _, ok := doc["scopes_supported"]; !ok {
		t.Fatal("missing scopes_supported")
	}
}

func TestRouterOAuthProtectedResource(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if doc["resource"] == nil || doc["authorization_servers"] == nil {
		t.Fatalf("incomplete PRM: %#v", doc)
	}
}

func TestHandleOAuthAuthorizationServer(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	rr := httptest.NewRecorder()

	h.handleOAuthAuthorizationServer(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if doc["issuer"] != "https://example.com" {
		t.Fatalf("issuer: got %#v", doc["issuer"])
	}
	for _, key := range []string{
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"grant_types_supported",
		"response_types_supported",
	} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("missing %s", key)
		}
	}
	if doc["jwks_uri"] != "https://example.com/.well-known/jwks.json" {
		t.Fatalf("jwks_uri: %#v", doc["jwks_uri"])
	}
	agentAuth, ok := doc["agent_auth"].(map[string]any)
	if !ok {
		t.Fatalf("agent_auth: %#v", doc["agent_auth"])
	}
	if agentAuth["skill"] != "https://example.com/auth.md" {
		t.Fatalf("skill: %#v", agentAuth["skill"])
	}
	if agentAuth["register_uri"] != "https://example.com/api/users/guest" {
		t.Fatalf("register_uri: %#v", agentAuth["register_uri"])
	}
	types, _ := agentAuth["identity_types_supported"].([]any)
	if len(types) == 0 || types[0] != "anonymous" {
		t.Fatalf("identity_types_supported: %#v", agentAuth["identity_types_supported"])
	}
	anon, ok := agentAuth["anonymous"].(map[string]any)
	if !ok {
		t.Fatalf("anonymous: %#v", agentAuth["anonymous"])
	}
	if anon["claim_uri"] != "https://example.com/api/users/register" {
		t.Fatalf("claim_uri: %#v", anon["claim_uri"])
	}
}

func TestHandleOpenIDConfiguration(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rr := httptest.NewRecorder()

	h.handleOpenIDConfiguration(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	for _, key := range []string{
		"issuer",
		"authorization_endpoint",
		"token_endpoint",
		"jwks_uri",
		"grant_types_supported",
		"response_types_supported",
	} {
		if _, ok := doc[key]; !ok {
			t.Fatalf("missing %s in openid-configuration", key)
		}
	}
}

func TestHandleJWKS(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rr := httptest.NewRecorder()

	h.handleJWKS(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	keys, ok := doc["keys"].([]any)
	if !ok {
		t.Fatalf("keys: %#v", doc["keys"])
	}
	if len(keys) != 0 {
		t.Fatalf("expected empty JWKS for HS256, got %d keys", len(keys))
	}
}

func TestRouterAuthMD(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/auth.md", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	if !strings.HasPrefix(rr.Body.String(), "# auth.md") {
		t.Fatalf("body: %q", rr.Body.String()[:min(40, rr.Body.Len())])
	}
}
