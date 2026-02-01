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

if [[ -z "${JWT_SECRET:-}" ]]; then
  JWT_SECRET="dev_jwt_secret"
fi

INVOKER_PORT="${INVOKER_PORT:-18079}"
POSTGRES_PORT="${POSTGRES_PORT:-55432}"
COUNTERSPELL_PORT="${COUNTERSPELL_PORT:-8710}"
RESET_DB="${RESET_DB:-1}"
STARTED_DOCKER=0
CS_PID=""

cleanup() {
  if [[ -n "${CS_PID}" ]] && kill -0 "${CS_PID}" >/dev/null 2>&1; then
    echo "Stopping counterspell (PID: ${CS_PID})..."
    kill "${CS_PID}" >/dev/null 2>&1 || true
  fi

  if [[ "$STARTED_DOCKER" -eq 1 ]]; then
    echo "Stopping docker compose services..."
    (cd "$INVOKER_DIR" && INVOKER_PORT="$INVOKER_PORT" POSTGRES_PORT="$POSTGRES_PORT" docker compose -f docker-compose.dev.yml down) || true
  fi
}

dump_invoker_logs() {
  if [[ "$STARTED_DOCKER" -eq 1 ]]; then
    echo ""
    echo "Invoker logs (last 200 lines):"
    (cd "$INVOKER_DIR" && docker compose -f docker-compose.dev.yml logs --no-color --tail=200 invoker) || true
    echo ""
  fi
}

trap cleanup EXIT

if command -v docker >/dev/null 2>&1; then
  if [[ "$RESET_DB" -eq 1 ]]; then
    echo "Resetting local invoker DB volume..."
    (cd "$INVOKER_DIR" && INVOKER_PORT="$INVOKER_PORT" POSTGRES_PORT="$POSTGRES_PORT" JWT_SECRET="$JWT_SECRET" docker compose -f docker-compose.dev.yml down -v) || true
  fi
  echo "Starting invoker + postgres via docker compose..."
  (cd "$INVOKER_DIR" && INVOKER_PORT="$INVOKER_PORT" POSTGRES_PORT="$POSTGRES_PORT" JWT_SECRET="$JWT_SECRET" docker compose -f docker-compose.dev.yml up -d --build)
  STARTED_DOCKER=1
else
  echo "Docker not found. Start invoker manually on port ${INVOKER_PORT}."
fi

echo "Waiting for invoker to be ready..."
invoker_ready=0
for _ in $(seq 1 60); do
  if curl -fsS --max-time 1 "http://localhost:${INVOKER_PORT}/ready" >/dev/null 2>&1; then
    invoker_ready=1
    break
  fi
  sleep 1
done

if [[ "$invoker_ready" -ne 1 ]]; then
  echo "Invoker did not become ready in time."
  dump_invoker_logs
  exit 1
fi

echo "Building counterspell..."
(cd "$ROOT_DIR" && make build)

LOG_DIR="$ROOT_DIR/tmp"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-handshake.log"
rm -f "$LOG_FILE"

INVOKER_BASE_URL="http://localhost:${INVOKER_PORT}"
OAUTH_REDIRECT_URI="http://localhost:${INVOKER_PORT}/api/v1/auth/callback"

echo "Starting counterspell (logs: $LOG_FILE)..."
ENV=dev INVOKER_BASE_URL="$INVOKER_BASE_URL" OAUTH_REDIRECT_URI="$OAUTH_REDIRECT_URI" \
  "$ROOT_DIR/counterspell" -addr ":${COUNTERSPELL_PORT}" 2>&1 | tee "$LOG_FILE" &
CS_PID=$!

echo ""
echo "Waiting for auth URL..."
auth_url=""
for _ in $(seq 1 120); do
  auth_url=$(grep -o 'auth_url="[^"]\+"' "$LOG_FILE" | tail -n 1 | sed 's/auth_url="//;s/"//' || true)
  if [[ -n "$auth_url" ]]; then
    break
  fi
  if ! kill -0 "$CS_PID" >/dev/null 2>&1; then
    echo "Counterspell exited before auth URL was produced."
    dump_invoker_logs
    exit 1
  fi
  sleep 1
done

if [[ -z "$auth_url" ]]; then
  echo "Timed out waiting for auth URL."
  dump_invoker_logs
  exit 1
fi

echo ""
echo "Open this URL to authenticate:"
echo "$auth_url"
echo ""
echo "Supabase Redirect URLs must include:"
echo "  $OAUTH_REDIRECT_URI"
echo ""

echo "Waiting for Counterspell to authenticate..."
for _ in $(seq 1 300); do
  if grep -q "Authenticated" "$LOG_FILE"; then
    echo ""
    echo "✅ Authentication successful."
    exit 0
  fi
  if grep -q "Authentication failed" "$LOG_FILE"; then
    echo ""
    echo "❌ Authentication failed. Check $LOG_FILE."
    dump_invoker_logs
    exit 1
  fi
  if ! kill -0 "$CS_PID" >/dev/null 2>&1; then
    echo "Counterspell exited unexpectedly. Check $LOG_FILE."
    dump_invoker_logs
    exit 1
  fi
  sleep 1
done

echo "Timed out waiting for authentication. Check $LOG_FILE."
dump_invoker_logs
exit 1
