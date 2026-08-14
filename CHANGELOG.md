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
