package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleAgentSkillsIndex(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-skills/index.json", nil)
	rr := httptest.NewRecorder()

	h.handleAgentSkillsIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
	var doc map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("json: %v", err)
	}
	if doc["$schema"] != agentSkillsSchema {
		t.Fatalf("$schema: %#v", doc["$schema"])
	}
	skills, ok := doc["skills"].([]any)
	if !ok || len(skills) == 0 {
		t.Fatalf("skills: %#v", doc["skills"])
	}
	for _, raw := range skills {
		entry, _ := raw.(map[string]any)
		for _, key := range []string{"name", "type", "description", "url", "digest"} {
			if entry[key] == nil || entry[key] == "" {
				t.Fatalf("skill missing %s: %#v", key, entry)
			}
		}
		if entry["type"] != "skill-md" {
			t.Fatalf("type: %#v", entry["type"])
		}
		digest, _ := entry["digest"].(string)
		if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
			t.Fatalf("digest format: %q", digest)
		}
	}
}

func TestAgentSkillDigestMatchesBody(t *testing.T) {
	skills, err := loadAgentSkills()
	if err != nil {
		t.Fatal(err)
	}
	h := testHandler(t)
	for _, s := range skills {
		req := httptest.NewRequest(http.MethodGet, s.RelURL, nil)
		req.SetPathValue("name", s.Name)
		rr := httptest.NewRecorder()
		h.handleAgentSkillMD(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: status %d", s.Name, rr.Code)
		}
		sum := sha256.Sum256(rr.Body.Bytes())
		want := "sha256:" + hex.EncodeToString(sum[:])
		if want != s.Digest {
			t.Fatalf("%s digest mismatch: index=%s body=%s", s.Name, s.Digest, want)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
			t.Fatalf("%s Content-Type: %q", s.Name, ct)
		}
	}
}

func TestRouterAgentSkillsIndex(t *testing.T) {
	handler := testStaticRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-skills/index.json", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d body=%q", rr.Code, rr.Body.String())
	}
}
