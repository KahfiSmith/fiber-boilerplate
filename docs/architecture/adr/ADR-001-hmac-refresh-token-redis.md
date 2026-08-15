# ADR-001: HMAC-Hashed Refresh Tokens in Redis with Atomic Rotation

## Context

The backend issues long-lived refresh tokens to keep users signed in. Storing
them insecurely would let an attacker impersonate any session. Common options:

- Store the raw token in a database column — a database leak exposes every
  active session.
- Store the raw token in Redis keyed by user ID — same leak risk, and one user
  can hold multiple sessions.
- Store only a one-way hash, keyed by the token hash, in Redis.

Refresh tokens also need rotation: each use should issue a new token so a
stolen token is usable once. Rotation must be atomic — concurrent refreshes
with the same token must not both succeed, and reuse of an already-rotated
token must revoke the whole token family.

## Decision

- Refresh tokens are 32-byte random values, `base64.RawURLEncoding`.
- The raw token is **never stored**. Only its HMAC-SHA256 hash
  (`REFRESH_TOKEN_HMAC_KEY`) is persisted.
- Storage is **Redis**, keyed by the token hash:
  - `auth:refresh:active:<hash>` → session ID
  - `auth:refresh:used:<hash>` → session ID (rotated tokens)
  - `auth:session:<session_id>` → session metadata JSON
  - `auth:family:<family_id>` → all hashes in a token family
- Rotation is performed by a single Redis Lua script (`AtomicRotationLuaScript`
  in `src/modules/auth/refresh.repository.go`) that moves the hash from active
  to used, writes the new hash, and updates session metadata in one atomic step.
- Reuse of a used token returns `REFRESH_TOKEN_REUSED` and revokes the entire
  family.

## Consequences

- A Redis leak exposes only hashes, not usable tokens.
- Atomic rotation prevents race conditions and enables reuse detection.
- Revocation is family-wide: logging out or reusing one token kills all sibling
  sessions.
- Requires Redis at runtime; the session store is volatile by design (cleared
  on Redis restart), which is acceptable for this auth model.
- HMAC (not plain SHA-256) means the stored hash is keyed, so a brute-force on
  leaked hashes requires the secret key.

## Status

Accepted.
