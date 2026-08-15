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

const PATH_RE = /`((?:src|cmd|db|scripts)\/[A-Za-z0-9_/.\-()*]+)`/g;
const seen = new Set();

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

const FEATURES_DIR = join(ROOT, "docs/features");
const MODULES_DIR = join(ROOT, "src/modules");

if (exists(MODULES_DIR)) {
  for (const entry of readdirSync(MODULES_DIR)) {
    const full = join(MODULES_DIR, entry);
    const stat = statSync(full);
    if (!stat.isDirectory()) continue;
    const featureDoc = join(FEATURES_DIR, `${entry}.md`);
    if (!exists(featureDoc)) {
      errors.push(
        `Feature "${entry}" has no docs/features/${entry}.md. ` +
          `Create it from docs/features/_TEMPLATE.md before committing.`
      );
    }
  }
}

for (const w of warnings) console.warn(`⚠  ${w}`);
for (const e of errors) console.error(`✖  ${e}`);

if (errors.length > 0) {
  console.error(`\ndocs:check FAILED — ${errors.length} error(s), ${warnings.length} warning(s).`);
  console.error("Fix the docs or the code so docs stay accurate.");
  process.exit(1);
}

console.log(`docs:check OK — ${markdownFiles.length} files, ${warnings.length} warning(s).`);
