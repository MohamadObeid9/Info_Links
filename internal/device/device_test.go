package device

import "testing"

func TestClassifyUserAgent(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{
			name: "iphone",
			ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			want: "phone",
		},
		{
			name: "android phone",
			ua:   "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			want: "phone",
		},
		{
			name: "android tablet",
			ua:   "Mozilla/5.0 (Linux; Android 13; SM-X900) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			want: "laptop",
		},
		{
			name: "ipad",
			ua:   "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			want: "laptop",
		},
		{
			name: "desktop chrome",
			ua:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			want: "laptop",
		},
		{
			name: "desktop firefox",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			want: "laptop",
		},
		{
			name: "generic tablet keyword",
			ua:   "Mozilla/5.0 (Linux; Android 12; Tablet) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
			want: "laptop",
		},
		{
			name: "empty user agent",
			ua:   "",
			want: "laptop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyUserAgent(tt.ua); got != tt.want {
				t.Fatalf("ClassifyUserAgent(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}
