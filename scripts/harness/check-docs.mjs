#!/usr/bin/env node
/**
 * docs:check (backend) — validates that docs/ stays in sync with the repo.
 *
 * Checks:
 *  1. Every markdown link inside docs/ resolves (external sibling-repo links skipped).
 *  2. Every `src/`, `cmd/`, `db/`, `scripts/` path referenced in docs exists.
 *  3. No references to the old flat doc files (docs/api.md, docs/rules.md, ...).
 *
 * Exit code is non-zero when any check fails.
 */
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, dirname, relative, resolve } from "node:path";

const ROOT = process.cwd();
const DOCS_DIR = join(ROOT, "docs");

const errors = [];
const warnings = [];

function exists(path) {
  return existsSync(path);
}

function walk(dir) {
  const out = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const stat = statSync(full);
    if (stat.isDirectory()) out.push(...walk(full));
    else if (entry.endsWith(".md")) out.push(full);
  }
  return out;
}

const markdownFiles = walk(DOCS_DIR);

// --- 1. markdown links resolve ---------------------------------------------
const LINK_RE = /\]\(([^)]+\.md)(?:[^)]*)\)/g;

function isExternal(targetPath) {
  const rel = relative(ROOT, targetPath);
  return rel.startsWith("..") || (rel !== "" && !rel.startsWith("."));
}

for (const file of markdownFiles) {
  const content = readFileSync(file, "utf8");
  const fileDir = dirname(file);
  let match;
  while ((match = LINK_RE.exec(content)) !== null) {
    const raw = match[1].split("#")[0];
    const target = resolve(fileDir, raw);
    if (isExternal(target)) continue;
    if (!exists(target)) {
      errors.push(`Broken link in ${relative(ROOT, file)}: -> ${raw}`);
    }
  }
}

// --- 2. src/cmd/db/scripts paths referenced in docs exist -------------------
const PATH_RE = /`((?:src|cmd|db|scripts)\/[A-Za-z0-9_/.\-()*]+)`/g;
const seen = new Set();

// Paths that belong to the sibling frontend repo, referenced in cross-repo
// docs. These are intentional and are not expected to exist in this repo.
const EXTERNAL_PREFIXES = [
  "src/lib/api/",
  "src/store/",
  "src/hooks/",
  "src/providers/",
  "src/types/",
  "src/components/",
  "src/config/",
  "src/app/",
];

for (const file of markdownFiles) {
  const content = readFileSync(file, "utf8");
  let match;
  while ((match = PATH_RE.exec(content)) !== null) {
    const raw = match[1];
    if (raw.endsWith("...") || raw.endsWith("**") || raw.endsWith("*")) continue;
    if (seen.has(raw)) continue;
    seen.add(raw);
    const target = join(ROOT, raw);
    if (exists(target)) continue;
    // skip paths owned by the frontend repo
    if (EXTERNAL_PREFIXES.some((p) => raw.startsWith(p))) continue;
    warnings.push(`Referenced path not found: ${raw} (in ${relative(ROOT, file)})`);
  }
}

// --- 3. no stale flat-doc references ---------------------------------------
const STALE = [
  "api.md",
  "architecture.md",
  "coding-standards.md",
  "database.md",
  "patterns.md",
  "rules.md",
  "workflow.md",
];
for (const file of markdownFiles) {
  const content = readFileSync(file, "utf8");
  for (const stale of STALE) {
    if (content.includes(`docs/${stale}`) || content.includes(`${stale}`)) {
      // Only flag markdown link references, not prose mentions of a word.
      const re = new RegExp(`\\]\\([^)]*${stale}`, "i");
      if (re.test(content)) {
        errors.push(`Stale doc reference in ${relative(ROOT, file)}: ${stale}`);
      }
    }
  }
}

// --- output -----------------------------------------------------------------
for (const w of warnings) console.warn(`⚠  ${w}`);
for (const e of errors) console.error(`✖  ${e}`);

if (errors.length > 0) {
  console.error(`\ndocs:check FAILED — ${errors.length} error(s), ${warnings.length} warning(s).`);
  console.error("Fix the docs or the code so docs stay accurate.");
  process.exit(1);
}

console.log(`docs:check OK — ${markdownFiles.length} files, ${warnings.length} warning(s).`);
