package seo

import (
	"regexp"
	"strings"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

var accentMap = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ä", "a", "ã", "a",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ò", "o", "ó", "o", "ô", "o", "ö", "o", "õ", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"À", "a", "Á", "a", "Â", "a", "Ä", "a",
	"È", "e", "É", "e", "Ê", "e", "Ë", "e",
	"Ì", "i", "Í", "i", "Î", "i", "Ï", "i",
	"Ò", "o", "Ó", "o", "Ô", "o", "Ö", "o",
	"Ù", "u", "Ú", "u", "Û", "u", "Ü", "u",
	"Ç", "c",
)

// ProgramSlug normalizes a program name to a URL slug.
func ProgramSlug(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	s := accentMap.Replace(strings.ToLower(name))
	s = nonSlugChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
