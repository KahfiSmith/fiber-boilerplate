-- Add Google OAuth support: password_hash becomes nullable, add oauth columns.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_subject VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth_provider_subject
    ON users (oauth_provider, oauth_subject)
    WHERE oauth_provider IS NOT NULL AND oauth_subject IS NOT NULL;
