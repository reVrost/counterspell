#!/usr/bin/env bash
set -euo pipefail

root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${root}" ]]; then
  echo "verify: not inside a git repo" >&2
  exit 1
fi

cd "${root}"

mapfile -d '' files < <(
  {
    git diff --name-only -z --diff-filter=ACMRTUXB
    git diff --name-only -z --diff-filter=ACMRTUXB --cached
    git ls-files --others --exclude-standard -z
  } | sort -zu
)

if (( ${#files[@]} == 0 )); then
  echo "verify: no local changes detected; skipping."
  exit 0
fi

has_go=false
has_ui=false
has_sql=false

for file in "${files[@]}"; do
  case "${file}" in
    ui/*) has_ui=true ;;
    *.svelte|*.svelte.ts|*.svelte.js) has_ui=true ;;
    *.go|go.mod|go.sum) has_go=true ;;
    *.sql|sqlc.yaml) has_sql=true ;;
  esac
done

if ${has_sql}; then
  has_go=true
fi

echo "verify: changes detected (go=${has_go}, ui=${has_ui}, sql=${has_sql})"

if ! ${has_go} && ! ${has_ui} && ! ${has_sql}; then
  echo "verify: no relevant changes; skipping."
  exit 0
fi

if ${has_sql}; then
  echo "verify: running make sqlc"
  make sqlc
fi

if ${has_go} || ${has_ui}; then
  echo "verify: building frontend"
  (cd ui && npm run build)
fi

if ${has_go}; then
  echo "verify: running make test"
  make test
fi

if ${has_ui}; then
  echo "verify: running make test-e2e"
  make test-e2e
fi
