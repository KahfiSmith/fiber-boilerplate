#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENTRYPOINT="${SWAG_MAIN_FILE:-$ROOT_DIR/cmd/api/main.go}"
OUTPUT_DIR="${SWAG_OUTPUT_DIR:-$ROOT_DIR/docs/swagger}"

if ! command -v swag >/dev/null 2>&1; then
  echo "swag CLI is required. Install it with: go install github.com/swaggo/swag/cmd/swag@latest" >&2
  exit 1
fi

if [[ ! -f "$ENTRYPOINT" ]]; then
  echo "swagger entrypoint not found: $ENTRYPOINT" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

exec swag init -g "$ENTRYPOINT" -o "$OUTPUT_DIR" "$@"
