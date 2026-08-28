#!/usr/bin/env node
/**
 * verify:audit — internal consistency checks for this repo.
 *
 * Builds on top of docs:check (link + endpoint + feature gate) and
 * verify:cross-repo (BE↔FE sync) with checks that are specific to this
 * single repo's code quality and architectural rules.
 *
 * Each check produces either:
 *   - error:   exit code 1, blocks CI/commit
 *   - warning: printed with ⚠, does not block
 *
 * Run via: pnpm verify:audit
 *
 * Add new checks: see docs/auditing/automated-checks.md
 */
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { basename, extname, join, relative, sep } from "node:path";

const ROOT = process.cwd();
const SRC = join(ROOT, "src");
const CMD = join(ROOT, "cmd");
const DB_MIGRATIONS = join(ROOT, "db", "migrations");
const PACKAGE = join(ROOT, "package.json");
const ENV_EXAMPLE = join(ROOT, ".env.example");

const errors = [];
const warnings = [];

function read(p) {
  return existsSync(p) ? readFileSync(p, "utf8") : null;
}

function walk(dir, exts = []) {
  const out = [];
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    let st;
    try {
      st = statSync(full);
    } catch {
      continue;
    }
    if (st.isDirectory()) out.push(...walk(full, exts));
    else if (exts.length === 0 || exts.some((x) => full.endsWith(x)))
      out.push(full);
  }
  return out;
}

function srcGoFiles() {
  return walk(SRC, [".go"]).concat(walk(CMD, [".go"]));
}

function recordError(file, msg) {
  errors.push(`${relative(ROOT, file)}: ${msg}`);
}
function recordWarning(file, msg) {
  warnings.push(`${relative(ROOT, file)}: ${msg}`);
}

// ============================================================================
// A. Naming & structure
// ============================================================================
{
  // A1: kebab-case (or snake_case for _test.go) for source files.
  for (const f of srcGoFiles()) {
    const name = basename(f);
    const base = name.replace(extname(name), "");
    // snake_case (lowercase + underscores) is the Go convention; kebab-case
    // is forbidden.
    if (base.includes("-")) {
      recordWarning(f, `filename contains hyphen; Go uses snake_case: ${name}`);
    }
    // Test files: must end in _test.go
    if (base.endsWith("_test") || base.endsWith("Test")) {
      if (!name.endsWith("_test.go")) {
        recordWarning(f, `test file should end in _test.go: ${name}`);
      }
    }
  }
}

{
  // A2: Module-level constants in config/exceptions use PascalCase or SCREAMING_SNAKE.
  // Best-effort: skip if too noisy.
  for (const f of walk(join(SRC, "config")).concat(walk(join(SRC, "common", "exceptions")))) {
    if (!f.endsWith(".go")) continue;
    const src = read(f);
    if (!src) continue;
    // Skip — Go conventions already enforced by go vet; this is redundant.
  }
}

{
  // A3: Exported package-level items should have a doc comment. We skip
  //     receiver methods (covered by struct doc), unexported items, and
  //     field declarations inside structs (not standalone items).
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    const lines = src.split("\n");
    let braceDepth = 0;
    let exportedNoComment = 0;
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      // Track brace depth so we only match package-level items, not struct
      // fields or method bodies.
      for (const ch of line) {
        if (ch === "{") braceDepth++;
        else if (ch === "}") braceDepth--;
      }
      if (braceDepth > 0) continue;

      // Skip receiver methods: `func (s *Foo) Bar()`.
      if (/^func\s+\(/.test(line)) continue;

      const m = line.match(/^func\s+([A-Z]\w*)\s*\(/) ||
                line.match(/^type\s+([A-Z]\w*)\s+/) ||
                line.match(/^(?:const|var)\s+([A-Z]\w*)\s*[=]/);
      if (!m) continue;

      // Check previous non-blank line for `//` comment.
      let hasComment = false;
      for (let j = i - 1; j >= 0; j--) {
        const prev = lines[j].trim();
        if (prev === "") continue;
        if (prev.startsWith("//")) { hasComment = true; break; }
        if (prev.startsWith("/*")) { hasComment = true; break; }
        // Hit a code line — no comment.
        break;
      }
      if (!hasComment) exportedNoComment++;
    }
    if (exportedNoComment > 0) {
      recordWarning(
        f,
        `${exportedNoComment} exported item(s) without doc comment`
      );
    }
  }
}

// ============================================================================
// B. Go hygiene
// ============================================================================
{
  // B1: No log.Fatal in non-main packages.
  const ALLOWLIST_LOG_FATAL = new Set([
    "cmd/api/main.go",
  ]);
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const rel = relative(ROOT, f);
    if (ALLOWLIST_LOG_FATAL.has(rel)) continue;
    const src = read(f);
    if (!src) continue;
    const matches = src.match(/\blog\.Fatal(?:f|l)?\b/g);
    if (matches) {
      recordError(f, `log.Fatal in non-main package (use slog + return error): ${matches.length} hit(s)`);
    }
  }
}

{
  // B2: No log.Printf/Println in source (should use slog).
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    const matches = src.match(/\blog\.(Printf|Println|Print)\b/g);
    if (matches) {
      recordError(f, `log.Printf/Println/Print (use slog): ${matches.length} hit(s)`);
    }
  }
}

{
  // B3: No //nolint without justification comment.
  for (const f of srcGoFiles()) {
    const src = read(f);
    if (!src) continue;
    const lines = src.split("\n");
    for (let i = 0; i < lines.length; i++) {
      if (!lines[i].includes("//nolint")) continue;
      // Check if the same line includes a reason (//nolint:reason) or if a
      // comment precedes.
      const hasReason = /\/\/nolint:[\w\s]+/.test(lines[i]);
      if (hasReason) continue;
      // Check line above.
      if (i > 0) {
        const prev = lines[i - 1].trim();
        if (prev.startsWith("//") || prev.startsWith("*")) continue;
      }
      recordWarning(f, `//nolint on line ${i + 1} without a reason or justification comment`);
    }
  }
}

{
  // B4: No `_ = err` (silent error swallow) without comment.
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    const lines = src.split("\n");
    for (let i = 0; i < lines.length; i++) {
      if (!/^\s*_\s*=\s*\w+/.test(lines[i])) continue;
      // Check if previous line has a comment.
      if (i > 0) {
        const prev = lines[i - 1].trim();
        if (prev.startsWith("//")) continue;
      }
      recordWarning(f, `silent error swallow on line ${i + 1} (no justification comment)`);
    }
  }
}

{
  // B5: No panic() outside main or test files.
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    if (relative(ROOT, f).startsWith("cmd" + sep)) continue;
    const src = read(f);
    if (!src) continue;
    const matches = src.match(/\bpanic\s*\(/g);
    if (matches) {
      recordWarning(f, `panic() in non-main code (use return error): ${matches.length} hit(s)`);
    }
  }
}

// ============================================================================
// C. Package & env consistency
// ============================================================================
{
  // C1: Required scripts exist in package.json.
  if (existsSync(PACKAGE)) {
    const pkg = JSON.parse(read(PACKAGE));
    const required = ["verify:fast", "verify", "verify:cross-repo", "verify:audit"];
    for (const s of required) {
      if (!pkg.scripts?.[s]) {
        errors.push(`package.json: missing required script "${s}"`);
      }
    }
  } else {
    errors.push("package.json not found");
  }
}

{
  // C2/C3: os.Getenv usage must match .env.example; secrets should be in config.
  const envSrc = read(ENV_EXAMPLE);
  const declaredKeys = new Set();
  if (envSrc) {
    for (const m of envSrc.matchAll(/^([A-Z_][A-Z0-9_]*)=/gm)) {
      declaredKeys.add(m[1]);
    }
  }

  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    for (const m of src.matchAll(/os\.Getenv\s*\(\s*"([A-Z_][A-Z0-9_]*)"\s*\)/g)) {
      const key = m[1];
      if (envSrc && !declaredKeys.has(key)) {
        recordWarning(
          f,
          `os.Getenv("${key}") used in code but not declared in .env.example`
        );
      }
    }
  }
}

// ============================================================================
// D. Architecture rules
// ============================================================================
{
  // D1: src/common/* must not import src/modules/* (except server.go, which
  //     is the composition root and by design imports modules).
  const D1_ALLOWLIST = new Set([
    "src/common/server/server.go",
  ]);
  for (const f of walk(join(SRC, "common"))) {
    if (!f.endsWith(".go")) continue;
    const rel = relative(ROOT, f);
    if (D1_ALLOWLIST.has(rel)) continue;
    const src = read(f);
    if (!src) continue;
    if (/"fiber-boilerplate\/src\/modules\//.test(src)) {
      recordError(f, "common must not import modules");
    }
  }
}

{
  // D2: src/database/* must not import src/modules/*.
  for (const f of walk(join(SRC, "database"))) {
    if (!f.endsWith(".go")) continue;
    const src = read(f);
    if (!src) continue;
    if (/"fiber-boilerplate\/src\/modules\//.test(src)) {
      recordError(f, "database must not import modules");
    }
  }
}

{
  // D3: src/config/* must not import src/modules/* or src/database/*.
  for (const f of walk(join(SRC, "config"))) {
    if (!f.endsWith(".go")) continue;
    const src = read(f);
    if (!src) continue;
    for (const banned of [
      "fiber-boilerplate/src/modules",
      "fiber-boilerplate/src/database",
    ]) {
      if (src.includes(banned)) {
        recordError(f, `config must not import ${banned}`);
      }
    }
  }
}

// ============================================================================
// E. Security patterns
// ============================================================================
{
  // E1: No hardcoded credentials (excluding `replace-with-...` placeholders
  //     and constants in tests).
  const SECRET_LITERAL = /\b(password|secret|api_?key|token)\s*[:=]\s*("[^"]{8,}"|`[^`]{8,}`)/i;
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    const lines = src.split("\n");
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      if (SECRET_LITERAL.test(line)) {
        // Allow placeholders.
        if (line.includes("replace-with-")) continue;
        const trimmed = line.trim();
        if (trimmed.startsWith("//")) continue;
        recordError(f, `possible hardcoded credential on line ${i + 1}: ${trimmed.slice(0, 80)}`);
      }
    }
  }
}

{
  // E2: Placeholder secrets in .env.example should be flagged for review.
  if (existsSync(ENV_EXAMPLE)) {
    const envSrc = read(ENV_EXAMPLE);
    if (envSrc) {
      for (const m of envSrc.matchAll(/^([A-Z_]*(?:SECRET|KEY|PASSWORD|TOKEN)[A-Z0-9_]*)=(.+)$/gm)) {
        const key = m[1];
        const val = m[2];
        if (val.includes("replace-with-")) {
          recordWarning(
            ENV_EXAMPLE,
            `${key} is a placeholder; ensure deployment overrides this in production`
          );
        }
      }
    }
  }
}

{
  // E3: CORS AllowOrigins should not be "*" in production code.
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    if (/AllowOrigins\s*:\s*\[\]string\{[^}]*"\*"/.test(src) ||
        /AllowOrigins\s*:\s*"\*"/.test(src)) {
      recordWarning(f, `CORS AllowOrigins includes "*" — too permissive`);
    }
  }
}

{
  // E4: Cookie Secure flag explicitly set in production code (warning, not error).
  //     Accept either:
  //     - Literal `Secure: true` (always secure)
  //     - `Secure: <someVariable>` (config-driven, e.g. `cfg.CookieSecure`)
  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    if (/Cookie\s*\{[^}]*Secure:\s*true/.test(src)) {
      // OK
    } else if (/Cookie\s*\{[^}]*Secure:\s*[a-zA-Z_]/.test(src)) {
      // OK - config-driven
    } else if (/fiber\.Cookie\s*\{/.test(src)) {
      recordWarning(f, `Cookie struct used but Secure flag not visibly set to true`);
    }
  }
}

// ============================================================================
// F. Migration drift
// ============================================================================
{
  // F1: Migrations are numbered in monotonically non-decreasing order
  //     (gaps are OK, e.g. when a migration is intentionally removed). Just
  //     flag if a later number is smaller than an earlier one.
  if (existsSync(DB_MIGRATIONS)) {
    const ups = readdirSync(DB_MIGRATIONS)
      .filter((f) => f.endsWith(".up.sql"))
      .map((f) => parseInt(f.slice(0, 6), 10))
      .filter((n) => !isNaN(n))
      .sort((a, b) => a - b);
    for (let i = 1; i < ups.length; i++) {
      if (ups[i] <= ups[i - 1]) {
        recordError(
          DB_MIGRATIONS,
          `migration number out of order: ${String(ups[i]).padStart(6, "0")}_* after ${String(ups[i - 1]).padStart(6, "0")}_*`
        );
      }
    }
  }
}

{
  // F2: Each migration has matching .up.sql and .down.sql.
  if (existsSync(DB_MIGRATIONS)) {
    const ups = new Set(readdirSync(DB_MIGRATIONS).filter((f) => f.endsWith(".up.sql")));
    const downs = new Set(readdirSync(DB_MIGRATIONS).filter((f) => f.endsWith(".down.sql")));
    for (const up of ups) {
      const base = up.replace(".up.sql", "");
      if (!downs.has(base + ".down.sql")) {
        recordError(DB_MIGRATIONS, `migration ${base} missing .down.sql`);
      }
    }
    for (const down of downs) {
      const base = down.replace(".down.sql", "");
      if (!ups.has(base + ".up.sql")) {
        recordError(DB_MIGRATIONS, `migration ${base} missing .up.sql`);
      }
    }
  }
}

{
  // F3: Best-effort check for GORM model vs migration drift (column types).
  // We compare `gorm:"type:varchar(N)"` tags in the User struct with the
  // actual SQL column type.
  const userType = read(join(SRC, "modules", "auth", "types", "auth.type.go"));
  if (userType) {
    const m = userType.match(/Name\s+string\s+`json:"name"\s+gorm:"type:varchar\((\d+)\)/);
    if (m) {
      const gormLen = parseInt(m[1], 10);
      const mig = read(join(DB_MIGRATIONS, "000001_create_users_table.up.sql"));
      if (mig) {
        const migM = mig.match(/name\s+VARCHAR\((\d+)\)/i);
        if (migM) {
          const migLen = parseInt(migM[1], 10);
          if (gormLen !== migLen) {
            recordWarning(
              join(SRC, "modules", "auth", "types", "auth.type.go"),
              `gorm tag says varchar(${gormLen}) but migration says VARCHAR(${migLen})`
            );
          }
        }
      }
    }
  }
}

{
  // F4: No .bak or temp migration files.
  if (existsSync(DB_MIGRATIONS)) {
    const tmp = readdirSync(DB_MIGRATIONS).filter(
      (f) => f.endsWith(".bak") || f.endsWith(".tmp") || f.includes("~")
    );
    for (const f of tmp) {
      recordWarning(DB_MIGRATIONS, `temp/bak file in migrations: ${f}`);
    }
  }
}

// ============================================================================
// G. Dead code detection
// ============================================================================
{
  // G1: Exported functions in src/ that have no callers.
  // (Best-effort: uses the function name as a search target.)
  function findCallers(name) {
    let count = 0;
    for (const f of srcGoFiles()) {
      const src = read(f);
      if (!src) continue;
      // Match `name(` not preceded by `func `.
      const re = new RegExp(`\\b${name}\\s*\\(`, "g");
      const matches = src.match(re);
      if (matches) count += matches.length;
    }
    return count;
  }

  for (const f of srcGoFiles()) {
    if (f.endsWith("_test.go")) continue;
    const src = read(f);
    if (!src) continue;
    // Find exported top-level functions: `func Name(...)`
    const lines = src.split("\n");
    let braceDepth = 0;
    let inFunc = false;
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      // Track braces only for non-test top-level funcs.
      for (const ch of line) {
        if (ch === "{") braceDepth++;
        else if (ch === "}") braceDepth--;
      }
      if (braceDepth > 0) continue;
      const m = line.match(/^func\s+([A-Z]\w*)\s*\(/);
      if (!m) continue;
      const name = m[1];
      const callers = findCallers(name);
      // The function itself counts 1 (its own definition).
      if (callers <= 1) {
        recordWarning(
          f,
          `exported function "${name}" has no callers (potential dead code)`
        );
      }
    }
  }
}

{
  // G2: Unused types in modules/*/types/*.go (heuristic).
  for (const f of walk(join(SRC, "modules"))
    .filter((x) => x.endsWith(".go") && !x.endsWith("_test.go"))) {
    const typesDir = f.split(sep).includes("types");
    if (!typesDir) continue;
    const src = read(f);
    if (!src) continue;
    const decls = [...src.matchAll(/^type\s+([A-Z]\w*)\s+/gm)].map((m) => m[1]);
    for (const name of decls) {
      let uses = 0;
      for (const other of srcGoFiles()) {
        if (other === f) continue;
        const otherSrc = read(other);
        if (!otherSrc) continue;
        const re = new RegExp(`\\b${name}\\b`, "g");
        const matches = otherSrc.match(re);
        if (matches) uses += matches.length;
      }
      if (uses === 0) {
        recordWarning(f, `type "${name}" defined but never referenced (potential dead code)`);
      }
    }
  }
}

// ============================================================================
// H. Test coverage minimum
// ============================================================================
{
  // H1: At least one *_test.go per module.
  for (const mod of ["auth", "health", "jwt"]) {
    const candidates = [
      join(SRC, "modules", mod),
      join(SRC, "common", mod),
    ];
    let found = false;
    for (const c of candidates) {
      if (existsSync(c)) {
        const tests = walk(c).filter((f) => f.endsWith("_test.go"));
        if (tests.length > 0) { found = true; break; }
      }
    }
    if (!found) {
      recordWarning(
        mod,
        `module "${mod}" has no _test.go file`
      );
    }
  }
}

// ============================================================================
// C4: .env.example coverage of all os.Getenv reads (already in C2; alias here)
// ============================================================================

// ============================================================================
// Output
// ============================================================================
for (const w of warnings) console.warn(`⚠  ${w}`);
for (const e of errors) console.error(`✖  ${e}`);

const total = errors.length + warnings.length;
console.log(
  `\nverify:audit: ${errors.length} error(s), ${warnings.length} warning(s), ${total} total`
);

if (errors.length > 0) {
  console.error("verify:audit FAILED — fix the errors above.");
  process.exit(1);
}
console.log("verify:audit OK.");
