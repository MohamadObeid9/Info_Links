package seo

import (
	"encoding/json"
	"fmt"
	"strings"

	"infolinks-backend/internal/repository"
)

func buildCourseJSONLD(baseURL string, data *repository.CoursePageData, canonical string) string {
	items := make([]map[string]interface{}, 0, len(data.Links))
	for i, l := range data.Links {
		items = append(items, map[string]interface{}{
			"@type":    "ListItem",
			"position": i + 1,
			"item": map[string]interface{}{
				"@type": "WebPage",
				"name":  l.Label,
				"url":   l.URL,
			},
		})
	}
	payload := map[string]interface{}{
		"@context": "https://schema.org",
		"@graph": []map[string]interface{}{
			{
				"@type":            "Course",
				"name":             data.Name,
				"courseCode":       strings.ToUpper(data.Code),
				"url":              canonical,
				"provider": map[string]interface{}{
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
