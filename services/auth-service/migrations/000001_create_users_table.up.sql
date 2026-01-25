CREATE SCHEMA IF NOT EXISTS auth;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE auth.users (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email                 VARCHAR(255) NOT NULL UNIQUE,
    password_hash         TEXT NOT NULL,
    is_verified           BOOLEAN NOT NULL DEFAULT FALSE,
    email_verify_token    TEXT,
    email_verify_sent_at  TIMESTAMPTZ,
    role                  VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    status                VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'blocked', 'banned')),
    last_login_at         TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Оптимальные индексы
CREATE INDEX idx_users_email              ON auth.users(email);
CREATE INDEX idx_users_role                ON auth.users(role);
CREATE INDEX idx_users_status              ON auth.users(status);
CREATE INDEX idx_users_verify_token        ON auth.users(email_verify_token) WHERE email_verify_token IS NOT NULL;

-- Автообновление updated_at
CREATE OR REPLACE FUNCTION auth.trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.trigger_set_updated_at();