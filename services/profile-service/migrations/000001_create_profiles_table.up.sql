CREATE SCHEMA IF NOT EXISTS profile;

CREATE TABLE IF NOT EXISTS profile.profiles (
    user_id      UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    avatar_url   TEXT,
    bio          TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profile.profiles(user_id);