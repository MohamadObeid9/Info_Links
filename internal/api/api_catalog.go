package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// linksetHref is one target in an RFC 9264 linkset relation array.
type linksetHref struct {
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

// apiCatalogEntry is one API in the RFC 9727 linkset catalog.
type apiCatalogEntry struct {
	Anchor      string        `json:"anchor"`
	ServiceDesc []linksetHref `json:"service-desc"`
	ServiceDoc  []linksetHref `json:"service-doc"`
	Status      []linksetHref `json:"status,omitempty"`
}

type apiCatalogDocument struct {
	Linkset []apiCatalogEntry `json:"linkset"`
}

func (h *Handler) baseURL() string {
	return strings.TrimSuffix(strings.TrimSpace(h.siteBaseURL), "/")
}

func (h *Handler) absURL(path string) string {
	base := h.baseURL()
	if path == "" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if base == "" {
		return path
	}
	return base + path
}

// handleAPICatalog serves /.well-known/api-catalog (RFC 9727).
func (h *Handler) handleAPICatalog(w http.ResponseWriter, r *http.Request) {
	catalogURL := h.absURL("/.well-known/api-catalog")
	w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="api-catalog"`, catalogURL))
	w.Header().Set("Content-Type", `application/linkset+json; profile="https://www.rfc-editor.org/info/rfc9727"`)

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	doc := apiCatalogDocument{
		Linkset: []apiCatalogEntry{
			{
				Anchor: h.absURL("/api"),
				ServiceDesc: []linksetHref{
					{Href: h.absURL("/openapi.json"), Type: "application/openapi+json"},
				},
				ServiceDoc: []linksetHref{
					{Href: h.absURL("/api/docs"), Type: "text/markdown"},
				},
				Status: []linksetHref{
					{Href: h.absURL("/healthz"), Type: "application/json"},
				},
			},
		},
	}

	body, err := json.Marshal(doc)
	if err != nil {
		h.LoggerWithID(r).Error("api catalog marshal failed", "error", err)
		writeJSONError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// handleOpenAPI serves the machine-readable OpenAPI description.
func (h *Handler) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/openapi+json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

// handleAPIDocs serves human-oriented API documentation (markdown).
func (h *Handler) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	var b strings.Builder
	b.WriteString("# Info Links API\n\n")
	b.WriteString("Machine-readable OpenAPI: [")
	b.WriteString(base)
	b.WriteString("/openapi.json](")
	b.WriteString(base)
	b.WriteString("/openapi.json)\n\n")
	b.WriteString("API catalog (RFC 9727): [")
	b.WriteString(base)
	b.WriteString("/.well-known/api-catalog](")
	b.WriteString(base)
	b.WriteString("/.well-known/api-catalog)\n\n")
	b.WriteString("Health: [")
	b.WriteString(base)
	b.WriteString("/healthz](")
	b.WriteString(base)
	b.WriteString("/healthz)\n\n")
	b.WriteString("## Overview\n\n")
	b.WriteString("Info Links exposes a JSON HTTP API for the CNAM Liban student materials hub.\n\n")
	b.WriteString("- **Content** — `GET /api/content` returns the full program → course → link tree.\n")
	b.WriteString("- **Students** — guest bootstrap, register, and login with first name + last name + number (1–100). No passwords.\n")
	b.WriteString("- **Analytics** — page views, link clicks, search, and browse events (student JWT).\n")
	b.WriteString("- **Submissions** — reports, feedback, and contributions (registered students).\n")
	b.WriteString("- **Admin** — JWT from `POST /api/auth/login`; CRUD under `/api/admin/...`.\n\n")
	b.WriteString("## Auth\n\n")
	b.WriteString("| Audience | How |\n|---|---|\n")
	b.WriteString("| Student | `Authorization: Bearer <token>` from `/api/users/guest`, `/register`, or `/login` |\n")
	b.WriteString("| Admin | `Authorization: Bearer <token>` from `POST /api/auth/login` |\n\n")
	b.WriteString("User id on analytics and submissions is taken from the JWT, never from the request body.\n\n")
	b.WriteString("## Quick links\n\n")
	b.WriteString("- Directory JSON: [")
	b.WriteString(base)
	b.WriteString("/api](")
	b.WriteString(base)
	b.WriteString("/api)\n")
	b.WriteString("- OpenAPI: [")
	b.WriteString(base)
	b.WriteString("/openapi.json](")
	b.WriteString(base)
	b.WriteString("/openapi.json)\n")

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}
