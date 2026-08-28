-- Migration 000002 created auth_sessions, otp_challenges, and auth_rate_limits.
-- The codebase uses Redis for all session/OTP/rate-limit storage, so these
-- Postgres tables are unused. This migration safely drops them if present.
DROP TABLE IF EXISTS auth_sessions CASCADE;
DROP TABLE IF EXISTS otp_challenges CASCADE;
DROP TABLE IF EXISTS auth_rate_limits CASCADE;
