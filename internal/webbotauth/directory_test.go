package webbotauth_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"infolinks-backend/internal/webbotauth"
)

func TestDirectoryJWKSContainsEd25519Key(t *testing.T) {
	dir, err := webbotauth.NewDirectory("test-secret", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(dir.JWKS(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Keys) != 1 {
		t.Fatalf("keys: %d", len(doc.Keys))
	}
	k := doc.Keys[0]
	if k["kty"] != "OKP" || k["crv"] != "Ed25519" || k["x"] == "" || k["kid"] == "" {
		t.Fatalf("jwk: %#v", k)
	}
	if _, err := base64.RawURLEncoding.DecodeString(k["x"]); err != nil {
		t.Fatalf("x decode: %v", err)
	}
}

func TestServeDirectorySigned(t *testing.T) {
	dir, err := webbotauth.NewDirectory("test-secret", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/.well-known/http-message-signatures-directory", nil)
	req.Host = "example.com"
	rr := httptest.NewRecorder()
	dir.ServeDirectory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/http-message-signatures-directory+json") {
		t.Fatalf("Content-Type: %q", ct)
	}
	if rr.Header().Get("Signature") == "" || rr.Header().Get("Signature-Input") == "" {
		t.Fatalf("missing signature headers: %+v", rr.Header())
	}
	if !strings.Contains(rr.Header().Get("Signature-Input"), `tag="http-message-signatures-directory"`) {
		t.Fatalf("Signature-Input: %q", rr.Header().Get("Signature-Input"))
	}
	if !strings.Contains(rr.Header().Get("Signature-Input"), "keyid=") {
		t.Fatalf("Signature-Input missing keyid: %q", rr.Header().Get("Signature-Input"))
	}

	// Verify signature over reconstructed base.
	var doc struct {
		Keys []map[string]string `json:"keys"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &doc)
	pubBytes, _ := base64.RawURLEncoding.DecodeString(doc.Keys[0]["x"])
	sigHeader := rr.Header().Get("Signature")
	raw := strings.TrimPrefix(sigHeader, "sig1=:")
	raw = strings.TrimSuffix(raw, ":")
	sig, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("sig decode: %v", err)
	}
	params := strings.TrimPrefix(rr.Header().Get("Signature-Input"), "sig1=")
	contentDigest := rr.Header().Get("Content-Digest")
	base := "\"@authority\";req: example.com\n" +
		"\"content-digest\": " + contentDigest + "\n" +
		"\"@signature-params\": " + params
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), []byte(base), sig) {
		t.Fatal("directory signature did not verify")
	}
}

func TestSignRequestHeaders(t *testing.T) {
	dir, err := webbotauth.NewDirectory("test-secret", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://crawltest.example/path", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := dir.SignRequest(req); err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Signature-Agent"); got != `"https://example.com"` {
		t.Fatalf("Signature-Agent: %q", got)
	}
	si := req.Header.Get("Signature-Input")
	if !strings.Contains(si, `tag="web-bot-auth"`) || !strings.Contains(si, "signature-agent") {
		t.Fatalf("Signature-Input: %q", si)
	}
	if req.Header.Get("Signature") == "" {
		t.Fatal("missing Signature")
	}
}
