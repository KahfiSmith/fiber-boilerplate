# Automated Checks

This document inventories all checks that run **automatically** via
`pnpm verify:audit` (and the wider `pnpm verify:all` chain), explains what
each one verifies, and shows how to extend the script.

## Architecture of the verify chain

```
pnpm verify:fast
  ├── go build ./...             (compile all packages)
  ├── go vet ./...               (static analysis)
  ├── go test ./...              (unit + integration)
  └── pnpm docs:check            (link + endpoint + feature gate)

pnpm verify:risk                 (path-based risk classification)

pnpm verify:cross-repo           (BE↔FE endpoint + error code sync)

pnpm verify:audit                (NEW — internal consistency)
  ├── Naming & structure
  ├── Go hygiene
  ├── Package & env consistency
  ├── Architecture rules
  ├── Security patterns
  ├── Migration drift
  ├── Dead code detection
  └── Test coverage minimum

pnpm verify:all
  └── runs all of the above

pnpm audit:report
  └── generates .audit/report-YYYY-MM-DD.{json,md} from all of the above
```

## Existing scripts (do not reimplement)

| Script | Purpose | Run via |
|---|---|---|
| `scripts/harness/check-docs.mjs` | Link resolution, src path, endpoint sync, feature gate | `pnpm docs:check` |
| `scripts/harness/check-cross-repo.mjs` | BE↔FE endpoint + error code sync | `pnpm verify:cross-repo` |
| `scripts/harness/check-risk.mjs` | Path-based risk classification (low/medium/high) | `pnpm verify:risk` |
| `scripts/harness/check-self-audit.mjs` | **NEW** internal consistency | `pnpm verify:audit` |

## New checks added by `check-self-audit.mjs`

### A. Naming & structure

| Check | What it does | Default |
|---|---|---|
| `A1` | All files in `src/` use kebab-case (or `_test.go` for tests) | warning |
| `A2` | Module-level constants in `config/`, `exceptions/` use PascalCase | warning |
| `A3` | All exported items have a doc comment (best-effort, excludes tests) | warning |

### B. Go hygiene

| Check | What it does | Default |
|---|---|---|
| `B1` | No `log.Fatal` in non-`main` packages (use slog or return error) | error |
| `B2` | No `log.Printf` in source (use slog) | error |
| `B3` | No `//nolint` without justification comment | warning |
| `B4` | No `_ = err` without comment (silent error swallow) | warning |
| `B5` | No `panic(` outside `main` or test files | warning |

### C. Package & env consistency

| Check | What it does | Default |
|---|---|---|
| `C1` | `package.json` has `verify:fast`, `verify`, `verify:audit` scripts | error |
| `C2` | All `os.Getenv("FOO")` references are listed in `.env.example` | warning |
| `C3` | No env var matching `*_SECRET`, `*_KEY`, `*_PASSWORD` is read in non-secure context | warning |

### D. Architecture rules

| Check | What it does | Default |
|---|---|---|
| `D1` | `src/common/` must not import from `src/modules/` | error |
| `D2` | `src/database/` must not import from `src/modules/` | error |
| `D3` | `src/config/` must not import from `src/modules/` or `src/database/` | error |
| `D4` | `cmd/api/main.go` is the only place that wires everything | (manual) |

### E. Security patterns

| Check | What it does | Default |
|---|---|---|
| `E1` | No hardcoded `password`, `secret`, `api_key`, `token` literals (excluding `replace-with-...` placeholders) | error |
| `E2` | No `JWT_*` env var defaults to `replace-with-` in `.env.example` (warning) | warning |
| `E3` | CORS `AllowOrigins` should not be `"*"` in production code | warning |
| `E4` | Cookie `Secure` flag explicitly set (true in prod) | warning |

### F. Migration drift

| Check | What it does | Default |
|---|---|---|
| `F1` | All migrations in `db/migrations/` are numbered sequentially with no gaps > 1 | warning |
| `F2` | Each migration has matching `.up.sql` and `.down.sql` files | error |
| `F3` | GORM model field tags don't drift from migration column types (best-effort scan) | warning |
| `F4` | Migrations never edited after being applied (no .bak or temp files) | warning |

### G. Dead code detection

| Check | What it does | Default |
|---|---|---|
| `G1` | Exported functions in `src/` that have no callers anywhere | warning |
| `G2` | Unused types in `src/modules/<name>/types/*.go` files (heuristic) | warning |

### H. Test coverage minimum

| Check | What it does | Default |
|---|---|---|
| `H1` | At least one `*_test.go` file exists per module under `src/modules/` | warning |
| `H2` | Critical paths (auth login, refresh rotation) have integration tests | warning |
| `H3` | `go test -race` is mentioned in docs (run with `-race` flag) | ➖ (manual) |

> **Note:** Line-level coverage is not in scope here. Use `go test -cover`
> in CI if a percentage threshold is needed.

## Severity model

Each check produces either:

- **error** — fails the script, exit code 1, blocks CI / commit
- **warning** — printed with `⚠`, does not block (visible in `audit:report`)

Rationale: warnings are valuable signal but not always blocking — for
example, a legacy `_ = err` in a non-production code path should be
visible but not necessarily fail a release.

## Allow-lists (false positive mitigation)

Some patterns are legitimately used in specific places. The script
maintains small allow-lists:

```js
// In check-self-audit.mjs:
const ALLOWLIST_LOG_FATAL = [
  "cmd/api/main.go",  // main is the only acceptable location
];

const ALLOWLIST_GETENV = [
  // LOG_LEVEL and LOG_ENCODING are read directly by logger.New()
  // and don't need to be in env config struct.
  "src/common/logger/logger.go",
];
```

When adding a new check, expect to add 1-2 allow-list entries. Document
each one with a comment explaining why.

## How to add a new check

1. **Edit** `scripts/harness/check-self-audit.mjs`. Add a new section:

   ```js
   // --- I. My new category --------------------------------------------
   for (const f of walk(join(SRC, "..."))) {
     const src = read(f);
     if (src.includes("forbiddenPattern")) {
       errors.push(`Found forbidden pattern in ${f}`);
     }
   }
   ```

2. **Document** it in this file under the corresponding section. Include:
   - What it checks
   - Severity (error or warning)
   - Any allow-list entries needed

3. **Add an item** to `SELF_AUDIT.md` master checklist with
   `Automation: ✅` so the agent/human knows it's covered.

4. **Test** the script: `pnpm verify:audit` should still pass on the
   current state. If it fails because of an existing issue, decide:
   - Fix the issue (preferred)
   - Add to allow-list with explanation
   - Or lower severity to warning

5. **Commit** with a message like:
   ```
   feat(audit): add check for <description>
   ```

## Running

```bash
# Standalone
pnpm verify:audit

# Full chain (includes go build/vet/test, docs:check, verify:cross-repo, verify:audit)
pnpm verify:all

# Generate aggregated report
pnpm audit:report
```

## Extending to the frontend

The frontend repo has a parallel `check-self-audit.mjs` with TS-specific
checks. See
[`../../Frontend/nextjs-boilerplate/docs/auditing/automated-checks.md`](../../Frontend/nextjs-boilerplate/docs/auditing/automated-checks.md)
for the TS equivalent.
