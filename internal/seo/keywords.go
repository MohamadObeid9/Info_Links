package seo

import "strings"

// InstitutionKeywords for meta and intro copy.
var InstitutionKeywords = []string{
	"CNAM", "CNAM Liban", "CNAM Lebanon", "Le CNAM", "ISAE CNAM", "ISAE-CNAM", "ISAE Liban",
}

// ProgramKeywords for broader searches.
var ProgramKeywords = []string{
	"génie informatique", "informatique", "licence info", "license info", "master", "AISL", "IRSM", "info",
}

// MaterialKeywords for resource types.
var MaterialKeywords = []string{
	"matériaux", "ressources", "supports", "archives", "TD", "TP", "cours", "sessions", "séances",
	"vidéos", "videos", "examens", "exams", "partiels", "CC", "rattrapage",
}

// ContentSection defines an on-page h2 block.
type ContentSection struct {
	ID    string
	Title string
	Types []string // normalized content_type tokens
}

// ContentSections is the fixed section order on course pages.
var ContentSections = []ContentSection{
	{ID: "examens", Title: "Examens & partiels", Types: []string{"exams", "examens", "exam", "partiels", "cc", "rattrapage"}},
	{ID: "td", Title: "TD & TP", Types: []string{"td", "tp"}},
	{ID: "cours", Title: "Cours", Types: []string{"cours", "course"}},
	{ID: "sessions", Title: "Sessions & séances", Types: []string{"sessions", "session", "seances", "séances"}},
	{ID: "videos", Title: "Vidéos", Types: []string{"videos", "video", "vidéos", "vidéo"}},
}

// SectionOther is the fallback section id.
const SectionOther = "autres"

// ClassifyLinkSection returns the section ID for a link based on content_type.
func ClassifyLinkSection(contentType string) string {
	if contentType == "" {
		return SectionOther
	}
	for _, part := range stringsSplitComma(contentType) {
		part = normalizeTypeToken(part)
		for _, sec := range ContentSections {
			for _, t := range sec.Types {
				if part == t {
					return sec.ID
				}
			}
		}
	}
	return SectionOther
}

func stringsSplitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalizeTypeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = accentReplacer.Replace(s)
	return s
}

var accentReplacer = strings.NewReplacer("é", "e", "è", "e", "ê", "e", "à", "a", "â", "a", "ù", "u", "û", "u", "ô", "o", "î", "i", "ï", "i", "ç", "c")
