-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS otp_deliveries;
DROP TABLE IF EXISTS otps;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_auth_providers;
DROP TABLE IF EXISTS users;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
