package service

import "testing"

func TestInferServiceLinkTarget(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://wa.me/9611234567", "WhatsApp"},
		{"https://api.whatsapp.com/send?phone=9611234567", "WhatsApp"},
		{"whatsapp://send?phone=9611234567", "WhatsApp"},
		{"https://t.me/example", "Telegram"},
		{"https://example.com", "website"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := inferServiceLinkTarget(tt.url); got != tt.want {
				t.Fatalf("inferServiceLinkTarget(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestResolveServiceLinkTarget(t *testing.T) {
	if got := ResolveServiceLinkTarget("Telegram", "https://wa.me/1"); got != "Telegram" {
		t.Fatalf("explicit target = %q", got)
	}
	if got := ResolveServiceLinkTarget("website", "https://wa.me/1"); got != "website" {
		t.Fatalf("explicit website target = %q", got)
	}
	if got := ResolveServiceLinkTarget("", "https://wa.me/1"); got != "WhatsApp" {
		t.Fatalf("inferred target = %q", got)
	}
}

func TestNormalizeLinkTargetLabel(t *testing.T) {
	tests := map[string]string{
		"WhatsApp":      "WhatsApp",
		"open whatsapp": "WhatsApp",
		"website":       "website",
		"Open website":  "website",
		"telegram":      "telegram",
		"contact":       "",
		"link":          "",
	}
	for in, want := range tests {
		if got := normalizeLinkTargetLabel(in); got != want {
			t.Fatalf("normalizeLinkTargetLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
