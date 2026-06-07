package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	const (
		dbURL  = "postgres://user:pass@localhost:5432/testdb"
		secret = "test-jwt-secret"
	)

	defaultCORS := "http://localhost:8080,http://localhost:5173"

	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "loads required vars with defaults",
			env: map[string]string{
				"DATABASE_URL":         dbURL,
				"JWT_SECRET":           secret,
				"PORT":                 "",
				"APP_ENV":              "",
				"CORS_ALLOWED_ORIGINS": "",
			},
			want: Config{
				Port:               "8080",
				AppEnv:             "development",
				DatabaseURL:        dbURL,
				JWTSecret:          secret,
				CorsAllowedOrigins: defaultCORS,
			},
		},
		{
			name: "loads custom optional vars",
			env: map[string]string{
				"DATABASE_URL":         dbURL,
				"JWT_SECRET":           secret,
				"PORT":                 "3000",
				"APP_ENV":              "production",
				"CORS_ALLOWED_ORIGINS": "https://example.com",
			},
			want: Config{
				Port:               "3000",
				AppEnv:             "production",
				DatabaseURL:        dbURL,
				JWTSecret:          secret,
				CorsAllowedOrigins: "https://example.com",
			},
		},
		{
			name: "trims whitespace from env values",
			env: map[string]string{
				"DATABASE_URL": "  " + dbURL + "  ",
				"JWT_SECRET":   "  " + secret + "  ",
				"PORT":         " 9090 ",
			},
			want: Config{
				Port:               "9090",
				AppEnv:             "development",
				DatabaseURL:        dbURL,
				JWTSecret:          secret,
				CorsAllowedOrigins: defaultCORS,
			},
		},
		{
			name: "missing database url",
			env: map[string]string{
				"DATABASE_URL": "",
				"JWT_SECRET":   secret,
			},
			wantErr: "database_url is required",
		},
		{
			name: "missing jwt secret",
			env: map[string]string{
				"DATABASE_URL": dbURL,
				"JWT_SECRET":   "",
			},
			wantErr: "jwt_secret is required",
		},
		{
			name: "database url whitespace only",
			env: map[string]string{
				"DATABASE_URL": "   ",
				"JWT_SECRET":   secret,
			},
			wantErr: "database_url is required",
		},
		{
			name: "jwt secret whitespace only",
			env: map[string]string{
				"DATABASE_URL": dbURL,
				"JWT_SECRET":   "   ",
			},
			wantErr: "jwt_secret is required",
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
