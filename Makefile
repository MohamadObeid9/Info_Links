# Docker Compose always loads project `.env` for YAML interpolation.
# DATABASE_URL may contain `$` (e.g. `$hAf` in a password); that is not a
# Compose variable. Disable the default `.env` for interpolation — the app
# still gets secrets via env_file format: raw.
export COMPOSE_DISABLE_ENV_FILE := 1

.PHONY: up down watch build rebuild

up:
	docker compose up

watch:
	docker compose up --watch

build:
	docker compose build

rebuild:
	docker compose up --build

down:
	docker compose down
