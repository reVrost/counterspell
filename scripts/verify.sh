#!/usr/bin/env bash
set -euo pipefail

repo_root="$(jj root 2>/dev/null || true)"
if [[ -z "${repo_root}" ]]; then
  echo "verify: not inside a jj repo" >&2
  exit 1
fi

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${repo_root}"

diff_output="$(jj diff --name-only)" || {
  echo "verify: failed to read jj diff" >&2
  exit 1
}

mapfile -t files < <(printf "%s\n" "${diff_output}" | sort -u)

if [[ "${project_root}" != "${repo_root}" ]]; then
  filtered=()
  for file in "${files[@]}"; do
    if [[ "${file}" == counterspell/* ]]; then
      filtered+=( "${file#counterspell/}" )
    fi
  done
  files=("${filtered[@]}")
fi

cd "${project_root}"

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
