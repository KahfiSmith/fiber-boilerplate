import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, dirname, resolve } from "node:path";

const ROOT = process.cwd();
const FE_ROOT = join(ROOT, "..", "..", "Frontend", "nextjs-boilerplate");

const errors = [];
const warnings = [];

function read(path) {
  return existsSync(path) ? readFileSync(path, "utf8") : null;
}

// --- 1. Routes match FE endpoints ------------------------------------------
const beControllerFile = join(ROOT, "src/modules/auth/auth.controller.go");
const feEndpointsFile = join(FE_ROOT, "src/lib/api/endpoints.ts");

if (!existsSync(FE_ROOT)) {
  errors.push(
    `Frontend repo not found at ${FE_ROOT}. Cross-repo check requires the sibling layout.`
  );
} else {
  const be = read(beControllerFile);
  const fe = read(feEndpointsFile);

  if (!be) errors.push("BE auth.controller.go missing");
  if (!fe) errors.push("FE endpoints.ts missing");

  if (be && fe) {
    const beRoutes = new Set(
      [...be.matchAll(/\.(Post|Get|Delete|Put|Patch)\("([^"]+)"/g)]
        .map((m) => m[2])
        .filter((p) => p !== "User-Agent")
        .map((p) => `/api/v1/auth${p}`)
    );
    const feEndpoints = new Set(
      [...fe.matchAll(/"(\/api\/v1\/auth\/[a-z-]+)"/g)].map((m) => m[1])
    );

    const beOnly = [...beRoutes].filter((p) => !feEndpoints.has(p));
    const feOnly = [...feEndpoints].filter((p) => !beRoutes.has(p));

    if (beOnly.length) {
      errors.push(`Routes in BE but not in FE endpoints: ${beOnly.join(", ")}`);
    }
    if (feOnly.length) {
      errors.push(`Endpoints in FE but not in BE routes: ${feOnly.join(", ")}`);
    }
  }
}

const beSources = [
  join(ROOT, "src/common/exceptions/exceptions.go"),
  join(ROOT, "src/common/middleware/auth.middleware.go"),
  join(ROOT, "src/common/middleware/origin.middleware.go"),
  join(ROOT, "src/common/response/response.go"),
  join(ROOT, "src/modules/auth/auth.service.go"),
  join(ROOT, "src/modules/auth/auth.controller.go"),
];

const feCodesRaw = read(join(FE_ROOT, "src/lib/api/auth-error-codes.ts"));
if (!feCodesRaw) {
  errors.push("FE auth-error-codes.ts missing");
} else {
  const feCodes = new Set(
    [...feCodesRaw.matchAll(/([A-Z][A-Z_]+): "([A-Z_]+)"/g)].map((m) => m[2])
  );

  const beCodes = new Set();
  for (const f of beSources) {
    const src = read(f);
    if (!src) continue;
    for (const m of src.matchAll(/(?:Unauthorized|Forbidden|BadRequest|NotFound|Internal|TooManyRequests)\(\s*"([A-Z_]+)"/g)) {
      beCodes.add(m[1]);
    }
    for (const m of src.matchAll(/codeStr = "([A-Z_]+)"/g)) {
      beCodes.add(m[1]);
    }
  }

  const beUnknown = [...beCodes].filter((c) => !feCodes.has(c));
  if (beUnknown.length) {
    errors.push(`Backend codes not present in FE auth-error-codes.ts: ${beUnknown.join(", ")}`);
  }
}

function walkMd(dir) {
  const out = [];
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const st = statSync(full);
    if (st.isDirectory()) out.push(...walkMd(full));
    else if (entry.endsWith(".md")) out.push(full);
  }
  return out;
}

const LINK_RE = /\]\(([^)]+\.md)(?:[^)]*)\)/g;
for (const file of walkMd(join(ROOT, "docs"))) {
  const content = readFileSync(file, "utf8");
  const dir = dirname(file);
  let m;
  while ((m = LINK_RE.exec(content)) !== null) {
    const raw = m[1].split("#")[0];
    const target = resolve(dir, raw);
    if (!target.startsWith(FE_ROOT)) continue;
    if (!existsSync(target)) {
      errors.push(`Cross-repo link broken in ${file}: -> ${raw}`);
    }
  }
}

for (const w of warnings) console.warn(`⚠  ${w}`);
for (const e of errors) console.error(`✖  ${e}`);

if (errors.length > 0) {
  console.error(`\nverify:cross-repo FAILED — ${errors.length} error(s), ${warnings.length} warning(s).`);
  process.exit(1);
}

console.log(`verify:cross-repo OK — BE↔FE in sync (${warnings.length} warning(s)).`);
