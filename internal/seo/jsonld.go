package seo

import (
	"encoding/json"
	"fmt"
	"strings"

	"infolinks-backend/internal/repository"
)

func buildCourseJSONLD(baseURL string, data *repository.CoursePageData, canonical string) string {
	items := make([]map[string]any, 0, len(data.Links))
	for i, l := range data.Links {
		items = append(items, map[string]any{
			"@type":    "ListItem",
			"position": i + 1,
			"item": map[string]any{
				"@type": "WebPage",
				"name":  l.Label,
				"url":   l.URL,
			},
		})
	}
	payload := map[string]any{
		"@context": "https://schema.org",
		"@graph": []map[string]any{
			{
				"@type":      "Course",
				"name":       data.Name,
				"courseCode": strings.ToUpper(data.Code),
				"url":        canonical,
				"provider": map[string]any{
					"@type": "Organization",
					"name":  "Le CNAM Liban — Info Links",
				},
				"description": BuildCourseDescription(data),
			},
			{
				"@type":           "ItemList",
				"itemListElement": items,
			},
		},
	}
	b, _ := json.Marshal(payload)
	return fmt.Sprintf(`<script type="application/ld+json">%s</script>`, b)
}
