DROP TRIGGER IF EXISTS set_updated_at ON auth.users;
DROP FUNCTION IF EXISTS auth.trigger_set_updated_at();
DROP TABLE IF EXISTS auth.users;
DROP SCHEMA IF EXISTS auth;
DROP EXTENSION IF EXISTS "uuid-ossp";