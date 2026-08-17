# Product Overview

## Product vision

This repository is the **backend foundation** for a web application that
requires authentication and per-user data. It serves the `nextjs-boilerplate`
frontend. The final product positioning, personas, and business workflows are
**not finalized yet** and must not be invented by the backend. The current
surface (auth + health) proves the session and API foundation so a real product
can be built on top of it.

For what is planned next, see the [Roadmap](../../ROADMAP.md).

## Product status

This repository is a **backend boilerplate**, not a finished product. The API
contract is documented in `docs/api/`; the implemented feature surface is in
`docs/features/`.

## Implemented product surface

- Authentication & user session (register, login, refresh, logout, password
  reset, email verification, delete account, current user).
- Health/readiness endpoint.

## Not implemented

- User personas and detailed workflows.
- Domain features beyond authentication (e.g. dashboards, content, billing).
- Email delivery (tokens are returned only; no mailer yet).

The current surface exists to prove the auth/session foundation only.
