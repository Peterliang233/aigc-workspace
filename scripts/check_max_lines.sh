#!/usr/bin/env bash
set -euo pipefail

MAX_LINES="${MAX_LINES:-200}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

fail=0

while IFS= read -r -d '' f; do
  # `wc -l` prefixes spaces; strip them.
  n="$(wc -l <"$f" | tr -d ' ')"
  if [[ "$n" -gt "$MAX_LINES" ]]; then
    printf "%s\t%s\n" "$n" "$f"
    fail=1
  fi
done < <(
  find . -type f \( -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.css' \) \
    -not -path './.git/*' \
    -not -path './frontend/node_modules/*' \
    -not -path './frontend/dist/*' \
    -not -path './backend/vendor/*' \
    -not -path './log/*' \
    -print0
)

if [[ "$fail" -ne 0 ]]; then
  echo
  echo "ERROR: some files exceed ${MAX_LINES} lines."
  echo "Fix by splitting code into smaller modules (recommended) or refactoring."
  exit 1
fi

echo "OK: all files are <= ${MAX_LINES} lines."

