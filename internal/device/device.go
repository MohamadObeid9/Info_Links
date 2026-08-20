package device

import "strings"

// ClassifyUserAgent maps a User-Agent string to "phone" or "laptop".
// phone: iPhone and Android phones (Mobile without iPad/Tablet).
// laptop: everything else, including desktops, iPads, and tablets.
func ClassifyUserAgent(ua string) string {
	lower := strings.ToLower(ua)

	if strings.Contains(lower, "iphone") {
		return "phone"
	}
	if strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet") {
		return "laptop"
	}
	if strings.Contains(lower, "android") && strings.Contains(lower, "mobile") {
		return "phone"
	}
	return "laptop"
}
