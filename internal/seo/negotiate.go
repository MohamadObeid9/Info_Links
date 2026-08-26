package seo

import (
	"net/http"
	"strconv"
	"strings"
)

// WantsMarkdown reports whether the client prefers a markdown representation
// via Accept: text/markdown (Markdown for Agents / content negotiation).
func WantsMarkdown(r *http.Request) bool {
	if r == nil {
		return false
	}
	return acceptPrefersMarkdown(r.Header.Get("Accept"))
}

func acceptPrefersMarkdown(accept string) bool {
	accept = strings.TrimSpace(accept)
	if accept == "" {
		return false
	}
	markdownQ := -1.0
	htmlQ := -1.0
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		media, q := parseAcceptPart(part)
		switch strings.ToLower(media) {
		case "text/markdown":
			markdownQ = q
		case "text/html":
			htmlQ = q
		}
	}
	if markdownQ < 0 {
		return false
	}
	if htmlQ < 0 {
		return markdownQ > 0
	}
	return markdownQ >= htmlQ && markdownQ > 0
}

func parseAcceptPart(part string) (media string, q float64) {
	q = 1
	pieces := strings.Split(part, ";")
	media = strings.TrimSpace(pieces[0])
	for _, p := range pieces[1:] {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "q=") {
			if v, err := strconv.ParseFloat(strings.TrimSpace(p[2:]), 64); err == nil {
				q = v
			}
		}
	}
	return media, q
}

func appendVary(w http.ResponseWriter, value string) {
	existing := w.Header().Get("Vary")
	if existing == "" {
		w.Header().Set("Vary", value)
		return
	}
	for _, part := range strings.Split(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(part), value) {
			return
		}
	}
	w.Header().Set("Vary", existing+", "+value)
}

// estimateTokens is a coarse chars/4 heuristic for the x-markdown-tokens header.
func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	n := (len(s) + 3) / 4
	if n < 1 {
		return 1
	}
	return n
}

func (h *Handler) writeHTML(w http.ResponseWriter, status int, payload string) {
	appendVary(w, "Accept")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(payload))
}

func (h *Handler) writeMarkdown(w http.ResponseWriter, status int, payload string) {
	appendVary(w, "Accept")
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("x-markdown-tokens", strconv.Itoa(estimateTokens(payload)))
	w.WriteHeader(status)
	_, _ = w.Write([]byte(payload))
}
