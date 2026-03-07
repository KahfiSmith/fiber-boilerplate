DROP INDEX IF EXISTS idx_auth_rate_limits_expires_at;
DROP TABLE IF EXISTS auth_rate_limits;

DROP INDEX IF EXISTS idx_otp_challenges_expires_at;
DROP INDEX IF EXISTS idx_otp_challenges_user_id;
DROP TABLE IF EXISTS otp_challenges;

DROP INDEX IF EXISTS idx_auth_sessions_expires_at;
DROP INDEX IF EXISTS idx_auth_sessions_user_id;
DROP INDEX IF EXISTS idx_auth_sessions_refresh_token_hash;
DROP TABLE IF EXISTS auth_sessions;
