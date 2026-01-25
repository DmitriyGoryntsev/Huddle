#!/bin/sh
set -e

sleep 2

# Ждем доступности базы данных
echo "⏳ Waiting for database ($DB_HOST:$DB_PORT) to start..."

until nc -z -v -w30 "$DB_HOST" "$DB_PORT"; do
  echo "Waiting for database connection..."
  sleep 2
done

echo "🟢 Database is up!"

# Формируем URL из универсальных переменных
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

echo "🔄 Running migrations for profile-service..."
# Запускаем мигратор
./bin/migrator -database "$DB_URL" -path migrations

echo "🚀 Starting profile-service..."
exec "$@"