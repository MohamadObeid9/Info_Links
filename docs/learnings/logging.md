## Logging in This Project

### Why use `slog` instead of `log`

`slog` is better for production backends because it supports structured logging with key/value fields.
That makes logs easier to filter in tools and easier to read during debugging.

```text
2026/04/28 13:44:10 user login 42
time=2026-04-28T13:44:10.123+03:00 level=INFO msg="user login" user_id=42 component=auth
```

- First line: plain text from `log`
- Second line: structured output from `slog`

### Log Levels

Use the four `slog` levels intentionally:

- `Debug`: detailed diagnostics for local development
- `Info`: expected lifecycle events (startup, successful connection)
- `Warn`: unexpected but recoverable behavior
- `Error`: operation failed and needs attention

### Viewing Logs (Text vs JSON)

You can emit logs in either text or JSON:

- Development: prefer text handler for readability in terminal
- Production: prefer JSON handler for log aggregation and filtering

Typical setup:

```go
if env == "development" {
	handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
} else {
	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
}
logger := slog.New(handler)
```

In hosted environments (Render, Datadog, ELK, Loki), JSON logs are easier to search by fields like `component`, `level`, and `error`.

### Where logging should happen

- Initialize the logger once in `main`.
- Pass the logger (or derived loggers) into packages.
- Return errors from package functions.
- Decide whether to exit or retry in `main` (application boundary).

This avoids duplicated logs and keeps control flow clear.

### `errors.New` vs `fmt.Errorf`

Use both, based on context:

- Use `errors.New("...")` for a fixed error with no underlying cause to wrap.
- Use `fmt.Errorf("...: %w", err)` when adding context to an existing error and preserving the original cause.

Examples:

```go
// No wrapped cause (validation/static error).
return errors.New("jwt secret is required")
```

```go
// Wrap lower-level error with context.
if err := db.Ping(); err != nil {
	return fmt.Errorf("ping database: %w", err)
}
```

### Quick Decision Guide

- Validation failed and there is no underlying cause -> use `errors.New("...")`
- You caught an error from another function and need context -> use `fmt.Errorf("...: %w", err)`
- Package-level function failed -> return `error` to caller
- Application boundary (`main`) decides process lifecycle -> log with `slog` and `os.Exit(1)` if needed
- Need to classify logs by subsystem -> use `logger.With("component", "...")`

### Error-Handling Pattern (Bad vs Good)

```go
// Bad: package code decides process lifecycle.
package example

import (
	"log/slog"
	"os"
)

func LogExample() {
	if _, err := someFunction(); err != nil {
		slog.Error("someFunction failed", "error", err)
		os.Exit(1)
	}
}
```

```go
// Good: package returns error; main decides what to do.
package example

import "fmt"

func LogExample() error {
	if _, err := someFunction(); err != nil {
		return fmt.Errorf("some function failed: %w", err)
	}
	return nil
}
```

```go
package main

import (
	"log/slog"
	"os"
)

func main() {
	if err := example.LogExample(); err != nil {
		slog.Error("log example failed", "component", "bootstrap", "error", err)
		os.Exit(1)
	}
}
```

### Quick Checklist

- Logger is configured once in `main`.
- Packages return `error` instead of calling `os.Exit`.
- Errors are wrapped with `%w` when propagating lower-level failures.
- Logs include stable fields like `component` and `error`.
- Production log level is `Info` (or higher); development can use `Debug`.
