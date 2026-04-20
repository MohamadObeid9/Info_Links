# InfoLinks Backend

This is the Go backend for InfoLinks, providing a REST API and statically serving the frontend.

## Architecture
The backend follows the Standard Go Project Layout:
- `cmd/server/main.go`: The entry point for the application.
- `internal/api/`: Contains handlers, routing, and HTTP logic.
- `internal/database/`: Contains the PostgreSQL connection setup.
- `internal/models/`: Contains the data structures that map to the database tables.

## Setup
1. Copy `.env.example` to `.env` (or set environment variables directly).
2. Set `DATABASE_URL` to your Postgres connection string (e.g., from Supabase).
3. Set `PORT` if you wish to run on a port other than `8080`.

## Testing
To run the full test suite (requires the `DATABASE_URL` environment variable pointing to a test database):
```bash
go test ./...
```

## Running Locally
To start the server:
```bash
go run cmd/server/main.go
```
The server will start on port `8080` by default and serve the frontend files located in `../frontend`.
