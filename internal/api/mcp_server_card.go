package api

import (
	"net/http"
)

// handleMCPServerCard serves /.well-known/mcp/server-card.json (SEP-1649).
func (h *Handler) handleMCPServerCard(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	doc := map[string]any{
		"$schema":         "https://static.modelcontextprotocol.io/schemas/mcp-server-card/v1.json",
		"version":         "1.0",
		"protocolVersion": "2025-06-18",
		"serverInfo": map[string]any{
			"name":        "info-links",
			"title":       "Info Links",
			"version":     "1.0.0",
			"description": "CNAM Liban student materials hub — browse courses and shared links via MCP tools backed by the Info Links HTTP API.",
			"homepage":    base + "/",
		},
		"description":       "Discover CNAM course materials, student auth, and analytics endpoints for Info Links.",
		"documentationUrl":  h.absURL("/api/docs"),
		"iconUrl":           h.absURL("/assets/android-chrome-192x192.png"),
		"transport": map[string]any{
			"type":     "streamable-http",
			"endpoint": h.absURL("/mcp"),
		},
		"capabilities": map[string]any{
			"tools":     map[string]any{"listChanged": false},
			"resources": map[string]any{"subscribe": false, "listChanged": false},
			"prompts":   map[string]any{"listChanged": false},
		},
		"authentication": map[string]any{
			"required": false,
			"schemes":  []string{"bearer"},
			"documentationUrl": h.absURL("/auth.md"),
		},
		"instructions": "Prefer public tools that call GET /api/content. For gated actions, follow /auth.md to obtain a student Bearer JWT and pass it as Authorization on MCP requests.",
		"tools":        "dynamic",
		"resources":    "dynamic",
		"prompts":      "dynamic",
	}
	writeDiscoveryJSON(w, doc)
}

// handleMCPEndpoint is a placeholder Streamable HTTP MCP entrypoint advertised
// by the server card. Full MCP method handling can be layered on later; for now
// clients can discover the card and learn the intended endpoint URL.
func (h *Handler) handleMCPEndpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost, http.MethodDelete:
		writeJSON(w, http.StatusNotImplemented, map[string]any{
			"error":             "MCP streamable HTTP transport is advertised but not fully implemented yet",
			"server_card":       h.absURL("/.well-known/mcp/server-card.json"),
			"documentation_url": h.absURL("/api/docs"),
			"auth":              h.absURL("/auth.md"),
		})
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
