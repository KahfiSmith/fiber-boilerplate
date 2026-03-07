CREATE TABLE IF NOT EXISTS auth_sessions (
    session_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(64) NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    ip_address VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_sessions_refresh_token_hash
    ON auth_sessions (refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id
    ON auth_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at
    ON auth_sessions (expires_at);

CREATE TABLE IF NOT EXISTS otp_challenges (
    challenge_id VARCHAR(64) PRIMARY KEY,
    purpose VARCHAR(64) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(64) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_otp_challenges_user_id
    ON otp_challenges (user_id);
CREATE INDEX IF NOT EXISTS idx_otp_challenges_expires_at
    ON otp_challenges (expires_at);

CREATE TABLE IF NOT EXISTS auth_rate_limits (
    rate_key VARCHAR(255) PRIMARY KEY,
    count BIGINT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_rate_limits_expires_at
    ON auth_rate_limits (expires_at);
