# Features Documentation

Overview of feature modules implemented in this backend.

## Implemented features

- [Authentication & User Session](./auth.md) - register, login,
  refresh, logout, password reset, email verification, delete account.
- [Health Check](./health.md) - DB + Redis health endpoint.

## Adding a new feature

1. Create the module under `src/modules/`.
2. Copy `docs/features/_TEMPLATE.md` to `docs/features/<module>.md`.
3. Fill in Overview, Core flow, Implementation map, Endpoints.
4. Add it to the list above.

The `docs:check` gate fails when a module has no feature doc. Feature docs
must describe only implemented behavior.
