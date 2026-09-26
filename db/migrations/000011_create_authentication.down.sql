DROP TABLE IF EXISTS agent_credential_scopes;
DROP TABLE IF EXISTS agent_credentials;
DROP TABLE IF EXISTS user_sessions;
DROP INDEX IF EXISTS users_normalized_email_unique_idx;
ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS password_hash;
