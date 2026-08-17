DROP INDEX IF EXISTS idx_users_oauth_provider_subject;

ALTER TABLE users DROP COLUMN IF EXISTS oauth_subject;
ALTER TABLE users DROP COLUMN IF EXISTS oauth_provider;

-- Re-apply the NOT NULL constraint to password_hash (best-effort for rollback).
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
