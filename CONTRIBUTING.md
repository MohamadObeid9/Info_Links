# Contributing

Thanks for helping improve Info Links — whether you contribute course resources, bug fixes, or backend improvements.

## Contribute resources (no code)

Use **Report** or **Contribute** in the live app at [infolinks.app](https://infolinks.app/).

## Contribute code

1. **Fork** the repository and create a branch:
   ```bash
   git checkout -b fix/your-change
   ```
2. **Set up** locally — see [README — Getting started](README.md#getting-started).
3. **Make your changes** and run checks:
   ```bash
   go test -race ./...
   golangci-lint run ./...   # if installed locally
   ```
4. **Open a pull request** with:
   - What changed and why
   - How you tested it
   - Screenshots for UI changes

## Guidelines

- Match existing code style and layering (`api` → `service` → `repository`)
- Add or update tests for backend logic changes
- For architectural choices, add or update an ADR in [`docs/adr/`](docs/adr/)
- Keep PRs focused — one concern per PR when possible
- Do not commit `.env` or secrets

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.
