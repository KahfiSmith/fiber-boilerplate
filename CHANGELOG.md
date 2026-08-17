# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- Restructured documentation into nested domain directories under `docs/` (api, architecture, conventions, database, development, infrastructure).
- Aligned documentation with the actual codebase (endpoints, error codes, env keys, folder structure, run modes).
- Documented the cross-repo relationship with the `nextjs-boilerplate` frontend.

### Added
- Verification harness: `package.json` with `verify:fast`/`verify`/`verify:all` scripts, risk classification, and cross-repo sync check.
- Pre-commit hook (`.githooks/pre-commit`) running `pnpm verify:fast`.
- GitHub Actions CI (`.github/workflows/ci.yml`): build, vet, test, docs check, risk classification.
- Feature documentation gate: new modules must be documented in `docs/features/<module>.md` (template enforced by `docs:check`).
- API collection + environment template under `docs/openapi/` for testing the API (Postman v2.1 format).
- Removed swagger artifacts (`docs/swagger.json`, `docs/swagger.yaml`, `scripts/swagger-generate.sh`) — the API collection under `docs/openapi/` replaces them as the importable contract.
- Added `ROADMAP.md` and `docs/product/overview.md`.
- Google SSO (OIDC, backend-first): `GET /auth/google` + `/auth/google/callback`, auto-create OAuth users, `password_hash` nullable + `oauth_provider`/`oauth_subject` columns (migration 000006). Disabled by default (`GOOGLE_ENABLED=false`).
