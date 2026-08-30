package service

import (
	"net/url"
	"strings"

	"infolinks-backend/internal/models"
)

// ResolveServiceLinkTarget returns an explicit target label or infers one from the URL.
func ResolveServiceLinkTarget(target, rawURL string) string {
	if label := normalizeLinkTargetLabel(target); label != "" {
		return label
	}
	return inferServiceLinkTarget(rawURL)
}

// NormalizeServiceClick fills in link_target and clicked URL before persistence.
func NormalizeServiceClick(click *models.ServiceClick) {
	click.URL = strings.TrimSpace(click.URL)
	if click.URL == "" {
		click.URL = strings.TrimSpace(click.ClickedURL)
	}
	click.LinkTarget = normalizeLinkTargetLabel(click.LinkTarget)
	if click.LinkTarget == "" {
		click.LinkTarget = inferServiceLinkTarget(click.URL)
	}
}

func normalizeLinkTargetLabel(target string) string {
	target = strings.TrimSpace(target)
	switch strings.ToLower(target) {
	case "", "contact", "link", "service":
		return ""
	case "whatsapp", "open whatsapp":
		return "WhatsApp"
	case "website", "open website":
		return "website"
	default:
		return target
	}
}

func inferServiceLinkTarget(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if u.Scheme == "whatsapp" || host == "wa.me" || host == "api.whatsapp.com" {
		return "WhatsApp"
	}
	if host == "t.me" || host == "telegram.me" || strings.HasSuffix(host, ".t.me") {
		return "Telegram"
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return "website"
	}
	return ""
}
