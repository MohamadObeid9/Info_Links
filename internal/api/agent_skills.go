package api

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed agent_skills
var agentSkillsFS embed.FS

const agentSkillsSchema = "https://schemas.agentskills.io/discovery/0.2.0/schema.json"

type agentSkillMeta struct {
	Name        string
	Description string
	RelURL      string
	Body        []byte
	Digest      string
}

func loadAgentSkills() ([]agentSkillMeta, error) {
	entries, err := fs.ReadDir(agentSkillsFS, "agent_skills")
	if err != nil {
		return nil, err
	}
	out := make([]agentSkillMeta, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		rel := path.Join("agent_skills", name, "SKILL.md")
		body, err := agentSkillsFS.ReadFile(rel)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", rel, err)
		}
		sum := sha256.Sum256(body)
		out = append(out, agentSkillMeta{
			Name:        name,
			Description: skillDescriptionFromMarkdown(body),
			RelURL:      "/.well-known/agent-skills/" + name + "/SKILL.md",
			Body:        body,
			Digest:      "sha256:" + hex.EncodeToString(sum[:]),
		})
	}
	return out, nil
}

func skillDescriptionFromMarkdown(body []byte) string {
	text := string(body)
	// Prefer YAML frontmatter description:
	if strings.HasPrefix(text, "---") {
		if end := strings.Index(text[3:], "---"); end >= 0 {
			front := text[3 : 3+end]
			for _, line := range strings.Split(front, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "description:") {
					return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
		}
	}
	return "Info Links agent skill"
}

// handleAgentSkillsIndex serves /.well-known/agent-skills/index.json (RFC draft v0.2.0).
func (h *Handler) handleAgentSkillsIndex(w http.ResponseWriter, r *http.Request) {
	skills, err := loadAgentSkills()
	if err != nil {
		h.LoggerWithID(r).Error("agent skills load failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	entries := make([]map[string]any, 0, len(skills))
	for _, s := range skills {
		entries = append(entries, map[string]any{
			"name":        s.Name,
			"type":        "skill-md",
			"description": s.Description,
			"url":         h.absURL(s.RelURL),
			"digest":      s.Digest,
		})
	}
	doc := map[string]any{
		"$schema": agentSkillsSchema,
		"skills":  entries,
	}
	writeDiscoveryJSON(w, doc)
}

// handleAgentSkillMD serves /.well-known/agent-skills/{name}/SKILL.md.
func (h *Handler) handleAgentSkillMD(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	rel := path.Join("agent_skills", name, "SKILL.md")
	body, err := agentSkillsFS.ReadFile(rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
