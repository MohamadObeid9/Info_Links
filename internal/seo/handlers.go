package seo

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"infolinks-backend/internal/errs"
	"infolinks-backend/internal/middleware"
	"infolinks-backend/internal/service"
)

// Handler serves server-rendered SEO pages.
type Handler struct {
	service *service.SEOService
	baseURL string
	logger  *slog.Logger
}

// NewHandler creates an SEO handler wired like other HTTP handlers in the project.
func NewHandler(logger *slog.Logger, seoService *service.SEOService, baseURL string) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		service: seoService,
		baseURL: strings.TrimSuffix(strings.TrimSpace(baseURL), "/"),
		logger:  logger,
	}
}

func (h *Handler) loggerWithID(r *http.Request) *slog.Logger {
	if id := middleware.RequestIDFromContext(r.Context()); id != "" {
		return h.logger.With("request_id", id)
	}
	return h.logger
}

func (h *Handler) writeHTML(w http.ResponseWriter, status int, payload string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(payload))
}

func (h *Handler) HandleCourse(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("code"))
	if code == "" {
		h.serve404(w)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.service.GetCoursePageByCode(ctx, code)
	if err != nil {
		if errors.Is(err, errs.ErrCourseNotFound) {
			h.serve404(w)
			return
		}
		h.loggerWithID(r).Error("course page failed", "error", err, "code", code)
		h.serve500HTML(w, r)
		return
	}
	html, err := renderCoursePage(h.baseURL, data)
	if err != nil {
		h.loggerWithID(r).Error("render course page failed", "error", err, "code", code)
		h.serve500HTML(w, r)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

func (h *Handler) HandleProgram(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.service.GetProgramBySlug(ctx, slug, ProgramSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.serve404(w)
			return
		}
		h.loggerWithID(r).Error("program page failed", "error", err, "slug", slug)
		h.serve500HTML(w, r)
		return
	}
	html, err := renderProgramPage(h.baseURL, data)
	if err != nil {
		h.loggerWithID(r).Error("render program page failed", "error", err, "slug", slug)
		h.serve500HTML(w, r)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

func (h *Handler) HandleCoursesIndex(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	entries, err := h.service.ListCoursesIndex(ctx)
	if err != nil {
		h.loggerWithID(r).Error("courses index failed", "error", err)
		h.serve500HTML(w, r)
		return
	}
	html, err := renderCoursesIndex(h.baseURL, entries)
	if err != nil {
		h.loggerWithID(r).Error("render courses index failed", "error", err)
		h.serve500HTML(w, r)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	writeSitemapURL(&b, h.baseURL+"/")
	writeSitemapURL(&b, h.baseURL+"/courses")

	codes, err := h.service.ListCourseCodesForSitemap(ctx)
	if err != nil {
		h.loggerWithID(r).Error("sitemap course codes failed", "error", err)
		h.serve500Plain(w, r)
		return
	}
	for _, code := range codes {
		writeSitemapURL(&b, h.baseURL+"/course/"+code)
	}

	programs, err := h.service.ListProgramsForSitemap(ctx, ProgramSlug)
	if err != nil {
		h.loggerWithID(r).Error("sitemap programs failed", "error", err)
		h.serve500Plain(w, r)
		return
	}
	for _, p := range programs {
		if p.Slug != "" {
			writeSitemapURL(&b, h.baseURL+"/program/"+p.Slug)
		}
	}
	b.WriteString(`</urlset>`)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(b.String()))
}

func writeSitemapURL(b *strings.Builder, loc string) {
	b.WriteString("<url><loc>")
	b.WriteString(xmlEscape(loc))
	b.WriteString("</loc></url>")
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func (h *Handler) HandleRobots(w http.ResponseWriter, r *http.Request) {
	body := strings.Join([]string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /admin",
		"Disallow: /admin-gate",
		"",
		"Sitemap: " + h.baseURL + "/sitemap.xml",
		"",
	}, "\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(body))
}

func (h *Handler) serve404(w http.ResponseWriter) {
	html, err := render404(h.baseURL)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	h.writeHTML(w, http.StatusNotFound, html)
}

func (h *Handler) serve500HTML(w http.ResponseWriter, r *http.Request) {
	id := middleware.RequestIDFromContext(r.Context())
	html, err := render500(h.baseURL, id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.writeHTML(w, http.StatusInternalServerError, html)
}

func (h *Handler) serve500Plain(w http.ResponseWriter, r *http.Request) {
	id := middleware.RequestIDFromContext(r.Context())
	msg := "Internal Server Error"
	if id != "" {
		msg += " (ref: " + id + ")"
	}
	http.Error(w, msg, http.StatusInternalServerError)
}
