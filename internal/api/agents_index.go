package api

import (
	"net/http"
)

// handleAgentsIndex serves /.well-known/agents-index.json — org-level DNS-AID /
// ANS-style directory listing A2A + MCP agents and their well-known cards.
func (h *Handler) handleAgentsIndex(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	doc := map[string]any{
		"origin":  "infolinks.app",
		"version": "1.0",
		"organization": map[string]any{
			"name": "Info Links",
			"url":  base + "/",
		},
		"agents": map[string]any{
			"info-links": map[string]any{
				"location": map[string]any{
					"fqdn":     "infolinks.app",
					"endpoint": h.absURL("/api"),
					"wellKnown": map[string]any{
						"a2a": h.absURL("/.well-known/agent-card.json"),
						"mcp": h.absURL("/.well-known/mcp/server-card.json"),
					},
				},
				"model-card": map[string]any{
					"description": "CNAM Liban student materials hub — browse courses, open shared links, and submit reports or contributions via a JSON HTTP API.",
					"version":     "1.0.0",
					"provider":    "Info Links",
				},
				"capability": map[string]any{
					"protocols": []string{"a2a", "mcp"},
				},
				"status": "ACTIVE",
			},
		},
	}
	writeDiscoveryJSON(w, doc)
}
