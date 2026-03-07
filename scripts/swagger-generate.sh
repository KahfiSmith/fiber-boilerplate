#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SEARCH_DIRS="${SWAG_SEARCH_DIRS:-cmd/api,pkg/controllers,pkg/dto/request,pkg/dto/response}"
ENTRYPOINT="${SWAG_MAIN_FILE:-main.go}"
OUTPUT_DIR="${SWAG_OUTPUT_DIR:-docs}"

if ! command -v swag >/dev/null 2>&1; then
  echo "swag CLI is required. Install it with: go install github.com/swaggo/swag/cmd/swag@latest" >&2
  exit 1
fi

cd "$ROOT_DIR"

FIRST_SEARCH_DIR="${SEARCH_DIRS%%,*}"
ENTRYPOINT_PATH="$ENTRYPOINT"

if [[ "$ENTRYPOINT_PATH" != */* ]]; then
  ENTRYPOINT_PATH="$FIRST_SEARCH_DIR/$ENTRYPOINT_PATH"
fi

if [[ ! -f "$ENTRYPOINT_PATH" ]]; then
  echo "swagger entrypoint not found: $ENTRYPOINT" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

exec swag init -g "$ENTRYPOINT" -d "$SEARCH_DIRS" -o "$OUTPUT_DIR" --outputTypes json,yaml "$@"
