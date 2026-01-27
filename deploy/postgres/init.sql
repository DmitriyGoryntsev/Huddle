-- Проверяем и создаем auth_db
SELECT 'CREATE DATABASE auth_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec

-- Проверяем и создаем profile_db
SELECT 'CREATE DATABASE profile_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'profile_db')\gexec