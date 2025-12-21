#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

usage() {
  cat <<'USAGE'
Usage:
  ./run.sh [--reset-db]

Options:
  --reset-db   Stop services and remove volumes (full DB wipe) before starting.
USAGE
}

reset_db=0
for arg in "$@"; do
  case "$arg" in
    --reset-db)
      reset_db=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      usage >&2
      exit 2
      ;;
  esac
done

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd docker

# docker compose can be a subcommand (docker compose) or a standalone binary (docker-compose)
compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    echo "Missing docker compose (install Docker Compose v2 or docker-compose)" >&2
    exit 1
  fi
}

if [[ ! -f .env ]]; then
  if [[ -f .env.example ]]; then
    cp .env.example .env
    echo "Created .env from .env.example"
  else
    echo "Missing .env and .env.example" >&2
    exit 1
  fi
fi

if [[ "$reset_db" -eq 1 ]]; then
  echo "Resetting DB (docker compose down -v)..."
  compose down -v
fi

# Start DB + API
compose up -d --build

# Wait for DB to accept connections
echo "Waiting for database..."
for i in {1..40}; do
  if compose exec -T db pg_isready -U telecombase -d telecombase >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
  if [[ $i -eq 40 ]]; then
    echo "Database is not ready" >&2
    exit 1
  fi
done

# Apply migrations (idempotent)
if [[ -f db/migrations/002_user_approval.sql ]]; then
  compose exec -T db psql -U telecombase -d telecombase < db/migrations/002_user_approval.sql >/dev/null
fi

# Health check
echo "Checking API health..."
for i in {1..40}; do
  if curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
    echo "API is up: http://localhost:8080"
    exit 0
  fi
  sleep 0.5
  if [[ $i -eq 40 ]]; then
    echo "API health check failed" >&2
    exit 1
  fi
done
