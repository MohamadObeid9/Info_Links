package seo

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"infolinks-backend/internal/repository"
)

type pageLayout struct {
	Title       string
	Description string
	Canonical   string
	BaseURL     string
	Body        template.HTML
	JSONLD      template.HTML
}

type coursePageView struct {
	Title       string
	Description string
	Canonical   string
	BaseURL     string
	Code        string
	Name        string
	Intro       string
	Placements  []repository.CoursePlacement
	Sections    []sectionView
	HasLinks    bool
	JSONLD      template.HTML
}

type sectionView struct {
	ID    string
	Title string
	Links []repository.SEOLink
}

type programPageView struct {
	Title       string
	Description string
	Canonical   string
	BaseURL     string
	ProgramName string
	Courses     []repository.ProgramCourseEntry
}

type coursesIndexView struct {
	Title       string
	Description string
	Canonical   string
	BaseURL     string
	Entries     []repository.CourseIndexEntry
}

func renderCoursePage(baseURL string, data *repository.CoursePageData) (string, error) {
	grouped := GroupLinksBySection(data.Links)
	var sections []sectionView
	for _, sec := range ContentSections {
		links := grouped[sec.ID]
		sections = append(sections, sectionView{ID: sec.ID, Title: sec.Title, Links: links})
	}
	sections = append(sections, sectionView{
		ID:    SectionOther,
		Title: "Autres ressources",
		Links: grouped[SectionOther],
	})

	intro := buildCourseIntro(data)
	title := BuildCourseTitle(data.Name, data.Code, data.Links)
	desc := BuildCourseDescription(data)
	canonical := fmt.Sprintf("%s/course/%s", strings.TrimSuffix(baseURL, "/"), strings.ToLower(data.Code))
	jsonld := template.HTML(buildCourseJSONLD(baseURL, data, canonical))

	view := coursePageView{
		Title:       title,
		Description: desc,
		Canonical:   canonical,
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		Code:        strings.ToUpper(data.Code),
		Name:        data.Name,
		Intro:       intro,
		Placements:  data.Placements,
		Sections:    sections,
		HasLinks:    len(data.Links) > 0,
		JSONLD:      jsonld,
	}
	return executeTemplate("course", view)
}

func buildCourseIntro(data *repository.CoursePageData) string {
	types := presentContentLabels(data.Links)
	typePhrase := "TD, cours, examens, sessions et vidéos"
	if len(types) > 0 {
		typePhrase = strings.Join(types, ", ")
	}
	return fmt.Sprintf(
		"Hub étudiant Info Links — matériaux CNAM Liban et ISAE CNAM pour %s (%s). "+
			"Trouvez %s : liens Drive, Telegram et Google Classroom partagés par les étudiants.",
		data.Name, strings.ToUpper(data.Code), typePhrase,
	)
}

func renderProgramPage(baseURL string, data *repository.ProgramPageData) (string, error) {
	title := fmt.Sprintf("%s — cours & matériaux | CNAM Liban Info Links", data.Name)
	desc := fmt.Sprintf(
		"Liste des cours %s au CNAM Liban — codes, TD, cours, examens, sessions. Hub Info Links pour génie informatique et licence.",
		data.Name,
	)
	canonical := fmt.Sprintf("%s/program/%s", strings.TrimSuffix(baseURL, "/"), data.Slug)
	view := programPageView{
		Title:       title,
		Description: desc,
		Canonical:   canonical,
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		ProgramName: data.Name,
		Courses:     data.Courses,
	}
	return executeTemplate("program", view)
}

func renderCoursesIndex(baseURL string, entries []repository.CourseIndexEntry) (string, error) {
	title := "Tous les cours CNAM — codes & matériaux | Info Links"
	desc := "Index des cours CNAM Liban (ISAE CNAM) : codes, TD, cours, examens, sessions, vidéos. Licence info, master, génie informatique."
	canonical := strings.TrimSuffix(baseURL, "/") + "/courses"
	view := coursesIndexView{
		Title:       title,
		Description: desc,
		Canonical:   canonical,
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		Entries:     entries,
	}
	return executeTemplate("coursesIndex", view)
}

func render404(baseURL string) (string, error) {
	body := `<main class="seo-main"><h1>Cours introuvable</h1><p>Ce code cours n'existe pas sur Info Links.</p><p><a href="/">Retour à l'accueil</a> · <a href="/courses">Tous les cours</a></p></main>`
	return executeLayout(pageLayout{
		Title:       "Cours introuvable | Info Links",
		Description: "Code cours non trouvé sur Info Links CNAM Liban.",
		Canonical:   strings.TrimSuffix(baseURL, "/") + "/",
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		Body:        template.HTML(body),
	})
}

var tplFuncs = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
}

func executeTemplate(name string, data interface{}) (string, error) {
	var bodyT *template.Template
	var err error
	switch name {
	case "course":
		bodyT, err = template.New("course").Funcs(tplFuncs).Parse(courseBodyTpl)
	case "program":
		bodyT, err = template.New("program").Funcs(tplFuncs).Parse(programBodyTpl)
	case "coursesIndex":
		bodyT, err = template.New("coursesIndex").Funcs(tplFuncs).Parse(coursesIndexBodyTpl)
	default:
		return "", fmt.Errorf("unknown template %s", name)
	}
	if err != nil {
		return "", err
	}
	var bodyBuf bytes.Buffer
	if err := bodyT.Execute(&bodyBuf, data); err != nil {
		return "", err
	}

	switch name {
	case "course":
		v := data.(coursePageView)
		return executeLayout(pageLayout{
			Title: v.Title, Description: v.Description, Canonical: v.Canonical,
			BaseURL: v.BaseURL, Body: template.HTML(bodyBuf.String()), JSONLD: v.JSONLD,
		})
	case "program":
		v := data.(programPageView)
		return executeLayout(pageLayout{
			Title: v.Title, Description: v.Description, Canonical: v.Canonical,
			BaseURL: v.BaseURL, Body: template.HTML(bodyBuf.String()),
		})
	case "coursesIndex":
		v := data.(coursesIndexView)
		return executeLayout(pageLayout{
			Title: v.Title, Description: v.Description, Canonical: v.Canonical,
			BaseURL: v.BaseURL, Body: template.HTML(bodyBuf.String()),
		})
	}
	return "", fmt.Errorf("unhandled template %s", name)
}

func executeLayout(p pageLayout) (string, error) {
	t, err := template.New("layout").Parse(layoutTpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const layoutTpl = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}"/>
<link rel="canonical" href="{{.Canonical}}"/>
<meta property="og:title" content="{{.Title}}"/>
<meta property="og:description" content="{{.Description}}"/>
<meta property="og:url" content="{{.Canonical}}"/>
<meta property="og:type" content="website"/>
<link rel="stylesheet" href="/styles/app.css"/>
<style>
.seo-main{max-width:900px;margin:0 auto;padding:24px 16px 48px;font-family:Inter,system-ui,sans-serif}
.seo-main h1{font-size:1.75rem;margin-bottom:.25rem}
.seo-code{color:var(--muted,#666);font-weight:600}
.seo-intro{line-height:1.6;margin:1rem 0 1.5rem;color:var(--text,#222)}
.seo-breadcrumb{font-size:.9rem;color:var(--muted,#666);margin:.5rem 0}
.seo-section{margin:1.5rem 0}
.seo-section h2{font-size:1.2rem;border-bottom:1px solid var(--border,#ddd);padding-bottom:.35rem}
.seo-links{list-style:none;padding:0;margin:.75rem 0}
.seo-links li{margin:.5rem 0}
.seo-links a{color:var(--accent,#6c63ff);font-weight:600}
.seo-empty{color:var(--muted,#888);font-style:italic}
.seo-cta{display:inline-block;margin-top:1.5rem;padding:.75rem 1.25rem;background:#6c63ff;color:#fff!important;border-radius:8px;text-decoration:none;font-weight:600}
.seo-faq{margin-top:2rem;padding:1rem;background:var(--surface,#f5f5f8);border-radius:8px}
.seo-table{width:100%;border-collapse:collapse;margin-top:1rem}
.seo-table th,.seo-table td{text-align:left;padding:.5rem;border-bottom:1px solid var(--border,#eee)}
.seo-nav{margin-bottom:1rem;font-size:.9rem}
</style>
{{if .JSONLD}}{{.JSONLD}}{{end}}
</head>
<body>
<nav class="seo-nav"><a href="/">Info Links</a> · <a href="/courses">Tous les cours</a></nav>
{{.Body}}
</body>
</html>`

const courseBodyTpl = `<main class="seo-main">
<h1>{{.Name}}</h1>
<p class="seo-code">{{.Code}}</p>
<p class="seo-intro">{{.Intro}}</p>
{{range .Placements}}
<p class="seo-breadcrumb">{{.ProgramName}} → {{.YearName}} → {{.SemesterName}}{{if .IsOptional}} (optionnel){{end}}</p>
{{end}}
{{range .Sections}}
<section class="seo-section" id="{{.ID}}">
<h2>{{.Title}}</h2>
{{if .Links}}
<ul class="seo-links">
{{range .Links}}
<li><a href="{{.URL}}" rel="noopener noreferrer">{{.Label}}</a>{{if .Note}} — <span>{{.Note}}</span>{{end}}{{if .ContentType}} <em>({{.ContentType}})</em>{{end}}</li>
{{end}}
</ul>
{{else}}
<p class="seo-empty">Aucun lien pour le moment.</p>
{{end}}
</section>
{{end}}
<div class="seo-faq">
<h2>FAQ</h2>
<p><strong>Où trouver les examens, TD et cours pour {{.Code}} ?</strong></p>
<p>Les sections ci-dessus regroupent les liens étudiants (Drive, Telegram, Classroom). Ouvrez l'application pour accéder rapidement à tous les liens.</p>
</div>
<a class="seo-cta" href="/?highlight={{lower .Code}}">Ouvrir dans Info Links</a>
</main>`

const programBodyTpl = `<main class="seo-main">
<h1>{{.ProgramName}}</h1>
<p class="seo-intro">Cours et matériaux CNAM Liban — TD, cours, examens, sessions, vidéos. Cliquez sur un code pour voir tous les liens.</p>
{{if .Courses}}
<table class="seo-table">
<thead><tr><th>Code</th><th>Cours</th></tr></thead>
<tbody>
{{range .Courses}}
<tr><td><a href="/course/{{.Code}}">{{upper .Code}}</a></td><td>{{.Name}}</td></tr>
{{end}}
</tbody>
</table>
{{else}}
<p class="seo-empty">Aucun cours listé.</p>
{{end}}
<a class="seo-cta" href="/">Ouvrir Info Links</a>
</main>`

const coursesIndexBodyTpl = `<main class="seo-main">
<h1>Tous les cours CNAM</h1>
<p class="seo-intro">Index Info Links — codes cours, matériaux, TD, cours, examens, sessions et vidéos pour CNAM Liban et ISAE CNAM (licence info, master, génie informatique).</p>
<table class="seo-table">
<thead><tr><th>Code</th><th>Cours</th><th>Programme</th></tr></thead>
<tbody>
{{range .Entries}}
<tr><td><a href="/course/{{.Code}}">{{upper .Code}}</a></td><td>{{.Name}}</td><td>{{.ProgramName}}</td></tr>
{{end}}
</tbody>
</table>
<a class="seo-cta" href="/">Ouvrir Info Links</a>
</main>`
