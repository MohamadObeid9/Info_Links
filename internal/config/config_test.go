package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	const (
		dbURL           = "postgres://user:pass@localhost:5432/testdb"
		localDbURL      = "postgres://postgres:postgres@localhost:5432/infolinks?sslmode=disable"
		secret          = "test-jwt-secret"
		supabseURL      = "https://random.supabase.co"
		supabaseAnonKey = "a-random-generated-key"
	)

	defaultCORS := "http://localhost:8080,http://localhost:5173"
	defaultSiteBaseURL := "http://localhost:8080"

	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "loads required vars with defaults",
			env: map[string]string{
				"DATABASE_URL":         localDbURL,
				"JWT_SECRET":           secret,
				"SUPABASE_URL":         supabseURL,
				"SUPABASE_ANON_KEY":    supabaseAnonKey,
				"PORT":                 "",
				"APP_ENV":              "",
				"LOG_LEVEL":            "",
				"CORS_ALLOWED_ORIGINS": "",
			},
			want: Config{
				Port:               "8080",
				AppEnv:             "development",
				LogLevel:           "debug",
				DatabaseURL:        localDbURL,
				JWTSecret:          secret,
				SupabaseURL:        supabseURL,
				SupabaseAnonKey:    supabaseAnonKey,
				CorsAllowedOrigins: defaultCORS,
				SiteBaseURL:        defaultSiteBaseURL,
			},
		},
		{
			name: "loads custom optional vars",
			env: map[string]string{
				"DATABASE_URL":                dbURL,
				"JWT_SECRET":                  secret,
				"SUPABASE_URL":                supabseURL,
				"SUPABASE_ANON_KEY":           supabaseAnonKey,
				"PORT":                        "3000",
				"APP_ENV":                     "production",
				"LOG_LEVEL":                   "info",
				"CORS_ALLOWED_ORIGINS":        "https://example.com",
				"METRICS_BASIC_AUTH_USER":     "grafana-scraper",
				"METRICS_BASIC_AUTH_PASSWORD": "metrics-secret",
			},
			want: Config{
				Port:                 "3000",
				AppEnv:               "production",
				LogLevel:             "info",
				DatabaseURL:          dbURL,
				JWTSecret:            secret,
				SupabaseURL:          supabseURL,
				SupabaseAnonKey:      supabaseAnonKey,
				CorsAllowedOrigins:   "https://example.com",
				SiteBaseURL:          defaultSiteBaseURL,
				MetricsBasicAuthUser: "grafana-scraper",
				MetricsBasicAuthPass: "metrics-secret",
			},
		},
		{
			name: "trims whitespace from env values",
			env: map[string]string{
				"DATABASE_URL":      "  " + localDbURL + "  ",
				"JWT_SECRET":        "  " + secret + "  ",
				"SUPABASE_URL":      "  " + supabseURL + "  ",
				"SUPABASE_ANON_KEY": "  " + supabaseAnonKey + "  ",
				"PORT":              " 9090 ",
			},
			want: Config{
				Port:               "9090",
				AppEnv:             "development",
				LogLevel:           "debug",
				DatabaseURL:        localDbURL,
				JWTSecret:          secret,
				SupabaseURL:        supabseURL,
				SupabaseAnonKey:    supabaseAnonKey,
				CorsAllowedOrigins: defaultCORS,
				SiteBaseURL:        defaultSiteBaseURL,
			},
		},
		{
			name: "loads custom site base url",
			env: map[string]string{
				"DATABASE_URL":      localDbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
				"SITE_BASE_URL":     "https://infolinks.example.com",
			},
			want: Config{
				Port:               "8080",
				AppEnv:             "development",
				LogLevel:           "debug",
				DatabaseURL:        localDbURL,
				JWTSecret:          secret,
				SupabaseURL:        supabseURL,
				SupabaseAnonKey:    supabaseAnonKey,
				CorsAllowedOrigins: defaultCORS,
				SiteBaseURL:        "https://infolinks.example.com",
			},
		},
		{
			name: "missing database url falls back to local in development",
			env: map[string]string{
				"DATABASE_URL":      "",
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			want: Config{
				Port:               "8080",
				AppEnv:             "development",
				LogLevel:           "debug",
				DatabaseURL:        localDbURL,
				JWTSecret:          secret,
				SupabaseURL:        supabseURL,
				SupabaseAnonKey:    supabaseAnonKey,
				CorsAllowedOrigins: defaultCORS,
				SiteBaseURL:        defaultSiteBaseURL,
			},
		},
		{
			name: "missing jwt secret",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        "",
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			wantErr: "jwt secret is required",
		},
		{
			name: "missing supabase url secret",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      "",
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			wantErr: "supabase url is required",
		},
		{
			name: "missing supabase anon key secret",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": "",
			},
			wantErr: "supabase anon key is required",
		},
		{
			name: "database url whitespace only falls back to local in development",
			env: map[string]string{
				"DATABASE_URL":      "   ",
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			want: Config{
				Port:               "8080",
				AppEnv:             "development",
				LogLevel:           "debug",
				DatabaseURL:        localDbURL,
				JWTSecret:          secret,
				SupabaseURL:        supabseURL,
				SupabaseAnonKey:    supabaseAnonKey,
				CorsAllowedOrigins: defaultCORS,
				SiteBaseURL:        defaultSiteBaseURL,
			},
		},
		{
			name: "jwt secret whitespace only",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        "   ",
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			wantErr: "jwt secret is required",
		},
		{
			name: "supabase url whitespace only",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      "   ",
				"SUPABASE_ANON_KEY": supabaseAnonKey,
			},
			wantErr: "supabase url is required",
		},
		{
			name: "supabase anon key whitespace only",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": "   ",
			},
			wantErr: "supabase anon key is required",
		},
		{
			name: "production requires metrics basic auth",
			env: map[string]string{
				"DATABASE_URL":      dbURL,
				"JWT_SECRET":        secret,
				"SUPABASE_URL":      supabseURL,
				"SUPABASE_ANON_KEY": supabaseAnonKey,
				"APP_ENV":           "production",
			},
			wantErr: "metrics basic auth is required in production",
		},
		{
			name: "metrics basic auth user without password",
			env: map[string]string{
				"DATABASE_URL":            dbURL,
				"JWT_SECRET":              secret,
				"SUPABASE_URL":            supabseURL,
				"SUPABASE_ANON_KEY":       supabaseAnonKey,
				"METRICS_BASIC_AUTH_USER": "grafana-scraper",
			},
			wantErr: "metrics basic auth user and password must both be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := Load()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Load() succeeded unexpectedly")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
