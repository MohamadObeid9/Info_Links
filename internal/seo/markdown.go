package seo

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"infolinks-backend/internal/repository"
)

func yamlFrontmatter(title, description string) string {
	var b strings.Builder
	b.WriteString("---\n")
	if title != "" {
		b.WriteString("title: ")
		b.WriteString(yamlScalar(title))
		b.WriteByte('\n')
	}
	if description != "" {
		b.WriteString("description: ")
		b.WriteString(yamlScalar(description))
		b.WriteByte('\n')
	}
	b.WriteString("---\n\n")
	return b.String()
}

func yamlScalar(s string) string {
	if s == "" {
		return `""`
	}
	if needsYAMLQuotes(s) {
		return strconv.Quote(s)
	}
	return s
}

func needsYAMLQuotes(s string) bool {
	if strings.TrimSpace(s) != s {
		return true
	}
	return strings.ContainsAny(s, ":#{}[]&*?|>!%@`'\",\n")
}

func renderCourseMarkdown(baseURL string, data *repository.CoursePageData) (string, error) {
	if data == nil {
		return "", fmt.Errorf("nil course data")
	}
	grouped := GroupLinksBySection(data.Links)
	title := BuildCourseTitle(data.Name, data.Code, data.Links)
	desc := BuildCourseDescription(data)
	canonical := fmt.Sprintf("%s/course/%s", strings.TrimSuffix(baseURL, "/"), strings.ToLower(data.Code))
	intro := buildCourseIntro(data)

	var b strings.Builder
	b.WriteString(yamlFrontmatter(title, desc))
	b.WriteString("# ")
	b.WriteString(data.Name)
	b.WriteString("\n\n")
	b.WriteString("**")
	b.WriteString(strings.ToUpper(data.Code))
	b.WriteString("**\n\n")
	b.WriteString(intro)
	b.WriteString("\n\n")

	for _, p := range data.Placements {
		b.WriteString("- ")
		b.WriteString(p.ProgramName)
		b.WriteString(" → ")
		b.WriteString(p.YearName)
		b.WriteString(" → ")
		b.WriteString(p.SemesterName)
		if p.IsOptional {
			b.WriteString(" (optionnel)")
		}
		b.WriteString("\n")
	}
	if len(data.Placements) > 0 {
		b.WriteByte('\n')
	}

	for _, sec := range ContentSections {
		writeMarkdownLinkSection(&b, sec.Title, grouped[sec.ID])
	}
	writeMarkdownLinkSection(&b, "Autres ressources", grouped[SectionOther])

	b.WriteString("## FAQ\n\n")
	b.WriteString("**Où trouver les examens, TD et cours pour ")
	b.WriteString(strings.ToUpper(data.Code))
	b.WriteString(" ?**\n\n")
	b.WriteString("Les sections ci-dessus regroupent les liens étudiants (Drive, Telegram, Classroom). ")
	b.WriteString("Ouvrez l'application pour accéder rapidement à tous les liens.\n\n")
	b.WriteString("[Ouvrir dans Info Links](")
	b.WriteString(strings.TrimSuffix(baseURL, "/"))
	b.WriteString("/?highlight=")
	b.WriteString(strings.ToLower(data.Code))
	b.WriteString(")\n")

	jsonld := strings.TrimSpace(buildCourseJSONLD(baseURL, data, canonical))
	if jsonld != "" {
		b.WriteString("\n```json\n")
		// strip <script> wrapper if present — emit raw JSON for agents
		b.WriteString(jsonLDBody(jsonld))
		b.WriteString("\n```\n")
	}
	return b.String(), nil
}

func writeMarkdownLinkSection(b *strings.Builder, title string, links []repository.SEOLink) {
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	if len(links) == 0 {
		b.WriteString("_Aucun lien pour le moment._\n\n")
		return
	}
	for _, link := range links {
		b.WriteString("- [")
		b.WriteString(link.Label)
		b.WriteString("](")
		b.WriteString(link.URL)
		b.WriteString(")")
		if link.Note != "" {
			b.WriteString(" — ")
			b.WriteString(link.Note)
		}
		if link.ContentType != "" {
			b.WriteString(" _(")
			b.WriteString(link.ContentType)
			b.WriteString(")_")
		}
		b.WriteString("\n")
	}
	b.WriteByte('\n')
}

func jsonLDBody(scriptTag string) string {
	s := strings.TrimSpace(scriptTag)
	const open = `<script type="application/ld+json">`
	const close = `</script>`
	if strings.HasPrefix(s, open) && strings.HasSuffix(s, close) {
		return strings.TrimSpace(s[len(open) : len(s)-len(close)])
	}
	return s
}

func renderProgramMarkdown(baseURL string, data *repository.ProgramPageData) (string, error) {
	if data == nil {
		return "", fmt.Errorf("nil program data")
	}
	title := fmt.Sprintf("%s — cours & matériaux | CNAM Liban Info Links", data.Name)
	desc := fmt.Sprintf(
		"Liste des cours %s au CNAM Liban — codes, TD, cours, examens, sessions. Hub Info Links pour génie informatique et licence.",
		data.Name,
	)
	base := strings.TrimSuffix(baseURL, "/")

	var b strings.Builder
	b.WriteString(yamlFrontmatter(title, desc))
	b.WriteString("# ")
	b.WriteString(data.Name)
	b.WriteString("\n\n")
	b.WriteString("Cours et matériaux CNAM Liban — TD, cours, examens, sessions, vidéos. Cliquez sur un code pour voir tous les liens.\n\n")
	if len(data.Courses) == 0 {
		b.WriteString("_Aucun cours listé._\n\n")
	} else {
		for _, c := range data.Courses {
			b.WriteString("- [")
			b.WriteString(strings.ToUpper(c.Code))
			b.WriteString("](")
			b.WriteString(base)
			b.WriteString("/course/")
			b.WriteString(strings.ToLower(c.Code))
			b.WriteString(") — ")
			b.WriteString(c.Name)
			b.WriteString("\n")
		}
		b.WriteByte('\n')
	}
	b.WriteString("[Ouvrir Info Links](")
	b.WriteString(base)
	b.WriteString("/)\n")
	return b.String(), nil
}

func renderCoursesIndexMarkdown(baseURL string, entries []repository.CourseIndexEntry) (string, error) {
	title := "Tous les cours CNAM — codes & matériaux | Info Links"
	desc := "Index des cours CNAM Liban (ISAE CNAM) : codes, TD, cours, examens, sessions, vidéos. Licence info, master, génie informatique."
	base := strings.TrimSuffix(baseURL, "/")

	var b strings.Builder
	b.WriteString(yamlFrontmatter(title, desc))
	b.WriteString("# Tous les cours CNAM\n\n")
	b.WriteString("Index Info Links — codes cours, matériaux, TD, cours, examens, sessions et vidéos pour CNAM Liban et ISAE CNAM (licence info, master, génie informatique).\n\n")
	for _, e := range entries {
		b.WriteString("- [")
		b.WriteString(strings.ToUpper(e.Code))
		b.WriteString("](")
		b.WriteString(base)
		b.WriteString("/course/")
		b.WriteString(strings.ToLower(e.Code))
		b.WriteString(") — ")
		b.WriteString(e.Name)
		if e.ProgramName != "" {
			b.WriteString(" (")
			b.WriteString(e.ProgramName)
			b.WriteString(")")
		}
		b.WriteString("\n")
	}
	b.WriteByte('\n')
	b.WriteString("[Ouvrir Info Links](")
	b.WriteString(base)
	b.WriteString("/)\n")
	return b.String(), nil
}

func render404Markdown(baseURL string) string {
	base := strings.TrimSuffix(baseURL, "/")
	var b strings.Builder
	b.WriteString(yamlFrontmatter("Cours introuvable | Info Links", "Code cours non trouvé sur Info Links CNAM Liban."))
	b.WriteString("# Cours introuvable\n\n")
	b.WriteString("Ce code cours n'existe pas sur Info Links.\n\n")
	b.WriteString("- [Retour à l'accueil](")
	b.WriteString(base)
	b.WriteString("/)\n")
	b.WriteString("- [Tous les cours](")
	b.WriteString(base)
	b.WriteString("/courses)\n")
	return b.String()
}

// ServeSPAMarkdown writes a markdown representation of SPA shell pages (/ and /about).
func ServeSPAMarkdown(w http.ResponseWriter, baseURL, path string) {
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	var md string
	switch path {
	case "/about":
		md = renderAboutMarkdown(base)
	default:
		md = renderHomeMarkdown(base)
	}
	appendVary(w, "Accept")
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("x-markdown-tokens", strconv.Itoa(estimateTokens(md)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md))
}

func renderHomeMarkdown(baseURL string) string {
	var b strings.Builder
	b.WriteString(yamlFrontmatter(
		"Info Links — matériaux de cours CNAM Liban",
		"Hub étudiant Info Links : TD, cours, examens, sessions et vidéos pour les programmes CNAM Liban et ISAE CNAM.",
	))
	b.WriteString("# Info Links\n\n")
	b.WriteString("Hub étudiant pour les matériaux de cours CNAM Liban (ISAE CNAM) — TD, cours, examens, sessions, vidéos et liens partagés.\n\n")
	b.WriteString("## Browse\n\n")
	b.WriteString("- [Tous les cours](")
	b.WriteString(baseURL)
	b.WriteString("/courses)\n")
	b.WriteString("- [À propos](")
	b.WriteString(baseURL)
	b.WriteString("/about)\n")
	b.WriteString("- [Signaler / contribuer](")
	b.WriteString(baseURL)
	b.WriteString("/report-submit)\n")
	b.WriteString("- [Feedback](")
	b.WriteString(baseURL)
	b.WriteString("/feedback)\n")
	return b.String()
}

func renderAboutMarkdown(baseURL string) string {
	var b strings.Builder
	b.WriteString(yamlFrontmatter(
		"À propos — Info Links",
		"Info Links est un hub étudiant pour trouver rapidement les matériaux de cours CNAM Liban.",
	))
	b.WriteString("# À propos\n\n")
	b.WriteString("Info Links regroupe les liens étudiants (Drive, Telegram, Classroom, etc.) pour les cours CNAM Liban et ISAE CNAM.\n\n")
	b.WriteString("Parcourez par programme, année et semestre, ou cherchez un code cours. Les étudiants inscrits peuvent ouvrir les liens, signaler des problèmes et contribuer de nouvelles ressources.\n\n")
	b.WriteString("- [Accueil](")
	b.WriteString(baseURL)
	b.WriteString("/)\n")
	b.WriteString("- [Tous les cours](")
	b.WriteString(baseURL)
	b.WriteString("/courses)\n")
	return b.String()
}
