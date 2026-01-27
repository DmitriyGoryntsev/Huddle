#!/bin/sh
set -e

BINARY=$1

echo "⏳ Waiting for PostgreSQL ($POSTGRES_HOST:$POSTGRES_PORT)..."
until nc -z "$POSTGRES_HOST" "$POSTGRES_PORT"; do
  sleep 1
done
echo "✅ PostgreSQL is ready!"

echo "⏳ Waiting for Redis ($REDIS_HOST:$REDIS_PORT)..."
until nc -z "$REDIS_HOST" "$REDIS_PORT"; do
  sleep 1
done
echo "✅ Redis is ready!"

DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DBNAME}?sslmode=${POSTGRES_SSLMODE}"

echo "📝 Running database migrations..."
/app/bin/migrator \
  -database "$DATABASE_URL" \
  -path /app/migrations \
  -command up

echo "🚀 Starting application..."
exec "$BINARY"
