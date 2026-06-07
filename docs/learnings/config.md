## Config in This Project

### Why centralize config

Centralizing config in one package gives you one source of truth for environment variables, defaults, and validation rules.

Benefits:

- Avoids duplicate `.env` loading across packages
- Keeps `main` startup flow simple and predictable
- Makes tests easier by injecting config values directly
- Reduces bugs from scattered `os.Getenv(...)` usage

### Core pattern

- `internal/config/config.go` owns reading env variables
- `Load()` returns a typed `Config` struct
- `main` calls `config.Load()` once at startup
- Other packages receive only what they need (for example, `DatabaseURL`)

### `.env` behavior in development vs production

Use `godotenv.Load()` as a best-effort helper:

```go
_ = godotenv.Load()
```

What this means:

- Local development: values from `.env` are loaded if file exists
- Production/Render: missing `.env` does not fail startup
- Required values still must exist in environment, or validation returns an error

### Recommended `Config` shape

Keep fields typed and idiomatic:

```go
type Config struct {
	Port               string
	AppEnv             string
	DatabaseURL        string
	JWTSecret          string
	CorsAllowedOrigins string
}
```

### Defaults and validation

Use small defaults only for non-sensitive values:

- Good defaults: `Port`, `AppEnv`
- No defaults for secrets: `DatabaseURL`, `JWTSecret`

Example:

```go
cfg := Config{
	Port:        getEnv("PORT", "8080"),
	AppEnv:      getEnv("APP_ENV", "development"),
	DatabaseURL: os.Getenv("DATABASE_URL"),
	JWTSecret:   os.Getenv("JWT_SECRET"),
}

if cfg.DatabaseURL == "" {
	return Config{}, fmt.Errorf("database_url is required")
}
if cfg.JWTSecret == "" {
	return Config{}, fmt.Errorf("jwt_secret is required")
}
```

### Where config should and should not be used

- `main`: load full config and wire dependencies
- `database`, `api`, services: receive needed values as params
- Avoid calling `os.Getenv` in many internal packages
- Avoid passing raw maps when a typed struct is clearer

### Quick Decision Guide

- Need environment variables for startup -> use `config.Load()`
- Need default value for optional env -> use helper like `getEnv(key, fallback)`
- Missing required secret -> return error from `Load()`
- Package needs one value only -> pass that value, not whole config
- Unsure where to call `godotenv.Load()` -> only at app entrypoint/config package

### Common mistakes to avoid

- Failing startup only because `.env` is missing in production
- Duplicating env loading in multiple packages
- Using non-idiomatic field names (`DatabaseUrl` instead of `DatabaseURL`)
- Storing runtime config structs inside domain models package

### Quick Checklist

- Config loads once at startup
- `.env` loading is best-effort
- Required secrets are validated
- Field names are idiomatic Go
- Internal packages do not own process-level config loading
