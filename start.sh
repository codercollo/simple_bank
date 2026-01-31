#!/bin/sh
set -e

echo "run db migration"

# Export environment variables from app.env
export $(grep -v '^#' /app/app.env | xargs)

/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"
