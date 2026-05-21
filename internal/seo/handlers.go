package seo

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"infolinks-backend/internal/repository"
)

// Handler serves server-rendered SEO pages.
type Handler struct {
	Repo    *repository.SEORepository
	BaseURL string
}

// NewHandler creates an SEO handler with SITE_BASE_URL from the environment.
func NewHandler() *Handler {
	base := strings.TrimSpace(os.Getenv("SITE_BASE_URL"))
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Handler{
		Repo:    repository.NewSEORepository(),
		BaseURL: strings.TrimSuffix(base, "/"),
	}
}

func (h *Handler) writeHTML(w http.ResponseWriter, status int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(html))
}

// HandleCourse serves GET /course/{code}
func (h *Handler) HandleCourse(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("code"))
	if code == "" {
		h.serve404(w)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.Repo.GetCoursePageByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			h.serve404(w)
			return
		}
		log.Println("SEO course page error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	html, err := renderCoursePage(h.BaseURL, data)
	if err != nil {
		log.Println("SEO render error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

// HandleProgram serves GET /program/{slug}
func (h *Handler) HandleProgram(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.Repo.GetProgramBySlug(ctx, slug, ProgramSlug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.serve404(w)
			return
		}
		log.Println("SEO program page error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	html, err := renderProgramPage(h.BaseURL, data)
	if err != nil {
		log.Println("SEO render error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

// HandleCoursesIndex serves GET /courses
func (h *Handler) HandleCoursesIndex(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	entries, err := h.Repo.ListCoursesIndex(ctx)
	if err != nil {
		log.Println("SEO courses index error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	html, err := renderCoursesIndex(h.BaseURL, entries)
	if err != nil {
		log.Println("SEO render error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.writeHTML(w, http.StatusOK, html)
}

// HandleSitemap serves GET /sitemap.xml
func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	base := h.BaseURL
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	writeSitemapURL(&b, base+"/")
	writeSitemapURL(&b, base+"/courses")

	codes, err := h.Repo.ListCourseCodesForSitemap(ctx)
	if err != nil {
		log.Println("SEO sitemap codes error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, code := range codes {
		writeSitemapURL(&b, base+"/course/"+code)
	}

	programs, err := h.Repo.ListProgramsForSitemap(ctx, ProgramSlug)
	if err != nil {
		log.Println("SEO sitemap programs error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, p := range programs {
		if p.Slug != "" {
			writeSitemapURL(&b, base+"/program/"+p.Slug)
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

// HandleRobots serves GET /robots.txt
func (h *Handler) HandleRobots(w http.ResponseWriter, r *http.Request) {
	body := strings.Join([]string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /admin",
		"Disallow: /admin-gate",
		"",
		"Sitemap: " + h.BaseURL + "/sitemap.xml",
		"",
	}, "\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(body))
}

func (h *Handler) serve404(w http.ResponseWriter) {
	html, err := render404(h.BaseURL)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	h.writeHTML(w, http.StatusNotFound, html)
}
