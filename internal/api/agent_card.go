package api

import (
	"net/http"
)

// handleAgentCard serves the A2A Agent Card at /.well-known/agent-card.json.
func (h *Handler) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	apiURL := h.absURL("/api")

	doc := map[string]any{
		"name":            "Info Links",
		"description":     "CNAM Liban student materials hub — browse courses, open shared links, and submit reports or contributions via a JSON HTTP API.",
		"version":         "1.0.0",
		"protocolVersion": "0.3",
		"url":             apiURL, // legacy field; prefer supportedInterfaces
		"documentationUrl": h.absURL("/api/docs"),
		"provider": map[string]any{
			"organization": "Info Links",
			"url":          base + "/",
		},
		"supportedInterfaces": []map[string]any{
			{
				"url":             apiURL,
				"protocolBinding": "HTTP+JSON",
				"protocolVersion": "1.0",
			},
		},
		"capabilities": map[string]any{
			"streaming":         false,
			"pushNotifications": false,
			"extendedAgentCard": false,
		},
		"defaultInputModes":  []string{"text/plain", "application/json"},
		"defaultOutputModes": []string{"application/json", "text/markdown", "text/html"},
		"securitySchemes": map[string]any{
			"studentBearer": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
				"description":  "Student session JWT from /api/users/guest, /register, or /login. See /auth.md.",
			},
		},
		"security": []map[string]any{
			{"studentBearer": []string{}},
		},
		"skills": []map[string]any{
			{
				"id":          "browse-courses",
				"name":        "Browse courses",
				"description": "Fetch the program → year → semester → course → link navigation tree via GET /api/content, or SEO pages under /courses and /course/{code}.",
				"tags":        []string{"content", "courses", "cnam"},
				"examples": []string{
					"List all courses for génie informatique",
					"Get materials for course NFA008",
				},
				"inputModes":  []string{"application/json", "text/plain"},
				"outputModes": []string{"application/json", "text/markdown", "text/html"},
			},
			{
				"id":          "student-session",
				"name":        "Student session",
				"description": "Bootstrap an anonymous guest JWT, claim it with name+number, or sign in. Required before gated link opens and submissions. Documented in /auth.md.",
				"tags":        []string{"auth", "student", "registration"},
				"examples": []string{
					"Create a guest session",
					"Register as ziad_baroudi_65",
				},
				"inputModes":  []string{"application/json"},
				"outputModes": []string{"application/json"},
			},
			{
				"id":          "record-activity",
				"name":        "Record activity",
				"description": "Post page views, link clicks, search, and browse events for analytics (student Bearer required).",
				"tags":        []string{"analytics", "telemetry"},
				"examples": []string{
					"Record a home page view",
					"Record opening Link 2 in JavaScript",
				},
				"inputModes":  []string{"application/json"},
				"outputModes": []string{"application/json"},
			},
			{
				"id":          "submit-feedback",
				"name":        "Submit feedback and reports",
				"description": "Registered students can POST reports, feedback, and link contributions.",
				"tags":        []string{"reports", "feedback", "contributions"},
				"examples": []string{
					"Report a broken Drive link",
					"Suggest a new exam PDF for a course",
				},
				"inputModes":  []string{"application/json"},
				"outputModes": []string{"application/json"},
			},
		},
	}

	writeDiscoveryJSON(w, doc)
}
