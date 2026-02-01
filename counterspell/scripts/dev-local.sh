#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVOKER_DIR="$ROOT_DIR/../invoker"

if [[ ! -d "$INVOKER_DIR" ]]; then
  echo "invoker repo not found next to counterspell. Expected: $INVOKER_DIR"
  exit 1
fi

if [[ -f "$INVOKER_DIR/.envrc" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$INVOKER_DIR/.envrc"
  set +a
else
  echo "Warning: $INVOKER_DIR/.envrc not found. Supabase envs may be missing."
fi

INVOKER_PORT="${INVOKER_PORT:-18079}"
POSTGRES_PORT="${POSTGRES_PORT:-55432}"

echo "Starting invoker + postgres via docker compose..."
(cd "$INVOKER_DIR" && INVOKER_PORT="$INVOKER_PORT" POSTGRES_PORT="$POSTGRES_PORT" docker compose -f docker-compose.dev.yml up -d --build)

echo ""
echo "Local invoker: http://localhost:${INVOKER_PORT}"
echo "Local postgres: localhost:${POSTGRES_PORT}"
echo ""
echo "Run Counterspell in another terminal with:"
echo "INVOKER_BASE_URL=http://localhost:${INVOKER_PORT} \\"
echo "OAUTH_REDIRECT_URI=http://localhost:${INVOKER_PORT}/api/v1/auth/callback \\"
echo "make dev"
echo ""
echo "Supabase Redirect URLs must include:"
echo "  http://localhost:${INVOKER_PORT}/api/v1/auth/callback"
