# Source from the repository root to load .env and normalize PostgreSQL settings.
set -a
. ./.env
set +a
if [ -z "${DATABASE_URL:-}" ] && [ -n "${DATABASE_URL_LOCAL:-}" ]; then
  DATABASE_URL=${DATABASE_URL_LOCAL#jdbc:}
fi
if [ -n "${DB_USER:-}" ]; then PGUSER=${PGUSER:-$DB_USER}; fi
if [ -n "${DB_PASSWORD:-}" ]; then PGPASSWORD=${PGPASSWORD:-$DB_PASSWORD}; fi
export DATABASE_URL PGUSER PGPASSWORD
