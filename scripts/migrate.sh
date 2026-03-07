#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"
MIGRATIONS_DIR="$ROOT_DIR/db/migrations"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # Support CRLF .env files created on Windows.
  # shellcheck disable=SC1090
  source <(sed 's/\r$//' "$ENV_FILE")
  set +a
fi

ACTION="up"
TARGET=""

case "${1:-}" in
  "")
    ;;
  up)
    TARGET="${2:-}"
    ;;
  down|status)
    ACTION="$1"
    TARGET="${2:-}"
    ;;
  help|-h|--help)
    ACTION="help"
    ;;
  *)
    TARGET="$1"
    ;;
esac

ensure_migrations_dir() {
  if [[ ! -d "$MIGRATIONS_DIR" ]]; then
    echo "folder migrations tidak ditemukan: $MIGRATIONS_DIR" >&2
    exit 1
  fi
}

ensure_psql() {
  if ! command -v psql >/dev/null 2>&1; then
    echo "psql is required to run migrations" >&2
    exit 1
  fi
}

generate_swagger() {
  if [[ "${GENERATE_SWAGGER_ON_MIGRATE:-true}" != "true" ]]; then
    return
  fi

  echo "generating swagger docs"
  "$ROOT_DIR/scripts/swagger-generate.sh"
}

psql_cmd() {
  if [[ -n "${DATABASE_URL:-}" ]]; then
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 "$@"
    return
  fi

  export PGPASSWORD="${DB_PASSWORD:-postgres}"
  export PGSSLMODE="${DB_SSLMODE:-disable}"
  export PGOPTIONS="-c TimeZone=${DB_TIMEZONE:-UTC}"

  psql \
    --host="${DB_HOST:-127.0.0.1}" \
    --port="${DB_PORT:-5432}" \
    --username="${DB_USER:-postgres}" \
    --dbname="${DB_NAME:-fiber_boilerplate}" \
    -v ON_ERROR_STOP=1 \
    "$@"
}

sorted_files() {
  local pattern="$1"
  find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name "$pattern" | sort
}

migration_version() {
  local file="$1"
  local name
  name="$(basename "$file")"
  name="${name%.up.sql}"
  name="${name%.down.sql}"
  printf '%s\n' "$name"
}

apply_sql_file() {
  local file="$1"
  echo "applying $(basename "$file")"
  psql_cmd -f "$file"
}

run_up() {
  local file
  local found=0

  if [[ -n "$TARGET" ]]; then
    file="$MIGRATIONS_DIR/$TARGET.up.sql"
    if [[ ! -f "$file" ]]; then
      echo "migration not found: $file" >&2
      exit 1
    fi
    apply_sql_file "$file"
    return
  fi

  while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    found=1
    apply_sql_file "$file"
  done < <(sorted_files "*.up.sql")

  if [[ "$found" -eq 0 ]]; then
    echo "tidak ada file migration .up.sql"
  fi
}

run_down() {
  local file

  if [[ -n "$TARGET" ]]; then
    file="$MIGRATIONS_DIR/$TARGET.down.sql"
    if [[ ! -f "$file" ]]; then
      echo "migration not found: $file" >&2
      exit 1
    fi
    apply_sql_file "$file"
    return
  fi

  file="$(sorted_files "*.down.sql" | tail -n 1)"
  if [[ -z "$file" ]]; then
    echo "tidak ada file migration .down.sql"
    return
  fi

  apply_sql_file "$file"
}

run_status() {
  local file
  local up_file
  local down_file

  printf "%-32s %-8s %-8s\n" "VERSION" "UP" "DOWN"

  while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    up_file="yes"
    down_file="no"

    if [[ -f "$MIGRATIONS_DIR/$(migration_version "$file").down.sql" ]]; then
      down_file="yes"
    fi

    printf "%-32s %-8s %-8s\n" "$(migration_version "$file")" "$up_file" "$down_file"
  done < <(sorted_files "*.up.sql")
}

print_usage() {
  cat <<'EOF'
usage:
  scripts/migrate.sh [version]
  scripts/migrate-down.sh [version]
  scripts/migrate-status.sh
EOF
}

ensure_migrations_dir

case "$ACTION" in
  status)
    run_status
    ;;
  up)
    ensure_psql
    generate_swagger
    run_up
    ;;
  down)
    ensure_psql
    generate_swagger
    run_down
    ;;
  help|-h|--help)
    print_usage
    ;;
  *)
    print_usage >&2
    exit 1
    ;;
esac
