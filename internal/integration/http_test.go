//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPReadyz(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	srv := httptest.NewServer(newTestRouter(t, dbClient))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /readyz status: got %d want 200", resp.StatusCode)
	}
}

func TestHTTPGuestAndMe(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	srv := httptest.NewServer(newTestRouter(t, dbClient))
	t.Cleanup(srv.Close)

	token := postGuestToken(t, srv.URL)

	resp := authedGet(t, srv.URL+"/api/users/me", token)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /api/users/me status: got %d body %s", resp.StatusCode, body)
	}

	var user map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if user["is_guest"] != true {
		t.Fatalf("is_guest: got %v want true", user["is_guest"])
	}
	if user["id"] == nil || user["id"] == float64(0) {
		t.Fatalf("expected guest id, got %#v", user["id"])
	}
}

func TestHTTPGuestRegisterClaimAndDuplicate(t *testing.T) {
	dbClient := openTestDB(t)
	resetDB(t, dbClient.DB)

	srv := httptest.NewServer(newTestRouter(t, dbClient))
	t.Cleanup(srv.Close)

	token := postGuestToken(t, srv.URL)

	registerBody := `{"first_name":"mohamad","last_name":"hassan","number":55}`
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/users/register", bytes.NewBufferString(registerBody))
	if err != nil {
		t.Fatalf("new register request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/users/register: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status: got %d body %s", resp.StatusCode, body)
	}

	var registerResp struct {
		Token string         `json:"token"`
		User  map[string]any `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registerResp.User["is_guest"] != false {
		t.Fatalf("registered user is_guest: got %v want false", registerResp.User["is_guest"])
	}
	if registerResp.User["handle"] != "mohamad_hassan_55" {
		t.Fatalf("handle: got %v want mohamad_hassan_55", registerResp.User["handle"])
	}

	meResp := authedGet(t, srv.URL+"/api/users/me", registerResp.Token)
	defer func() { _ = meResp.Body.Close() }()
	if meResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(meResp.Body)
		t.Fatalf("GET /api/users/me after register: got %d body %s", meResp.StatusCode, body)
	}

	dupReq, err := http.NewRequest(http.MethodPost, srv.URL+"/api/users/register", bytes.NewBufferString(registerBody))
	if err != nil {
		t.Fatalf("new duplicate register request: %v", err)
	}
	dupReq.Header.Set("Content-Type", "application/json")
	dupReq.Header.Set("Authorization", "Bearer "+postGuestToken(t, srv.URL))

	dupResp, err := http.DefaultClient.Do(dupReq)
	if err != nil {
		t.Fatalf("duplicate register: %v", err)
	}
	defer func() { _ = dupResp.Body.Close() }()
	if dupResp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(dupResp.Body)
		t.Fatalf("duplicate register status: got %d want 409 body %s", dupResp.StatusCode, body)
	}
}

func postGuestToken(t *testing.T, baseURL string) string {
	t.Helper()
	resp, err := http.Post(baseURL+"/api/users/guest", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/users/guest: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("guest status: got %d body %s", resp.StatusCode, body)
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode guest token: %v", err)
	}
	if payload.Token == "" {
		t.Fatal("guest token is empty")
	}
	return payload.Token
}

func authedGet(t *testing.T, url, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new GET %s: %v", url, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}
