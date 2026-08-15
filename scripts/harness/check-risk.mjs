import { execSync } from "node:child_process";

const ROOT = process.cwd();

const args = process.argv.slice(2);
const baseIdx = args.indexOf("--base");
const headIdx = args.indexOf("--head");
const base = baseIdx !== -1 ? args[baseIdx + 1] : "HEAD";
const head = headIdx !== -1 ? args[headIdx + 1] : null;

function gitChangedPaths() {
  const baseRev = head ? `${base}...${head}` : base;
  try {
    const committed = execSync(`git diff --name-only ${baseRev}`, {
      encoding: "utf8",
      cwd: ROOT,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    const staged = execSync(`git diff --cached --name-only`, {
      encoding: "utf8",
      cwd: ROOT,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    const untracked = execSync(`git ls-files --others --exclude-standard`, {
      encoding: "utf8",
      cwd: ROOT,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    const set = new Set(
      [committed, staged, untracked].filter(Boolean).flatMap((s) => s.split("\n")).filter(Boolean)
    );
    return [...set];
  } catch {
    return [];
  }
}

const RULES = [
  { risk: "high", test: (p) => /^src\/modules\/auth\//.test(p) },
  { risk: "high", test: (p) => /^src\/common\//.test(p) },
  { risk: "high", test: (p) => /^src\/(config|database)\//.test(p) },
  { risk: "high", test: (p) => /^(Dockerfile|docker-compose\.yml|\.env\.example|go\.mod|go\.sum)/.test(p) },
  { risk: "high", test: (p) => /^scripts\/harness\//.test(p) },
  { risk: "high", test: (p) => /^\.github\//.test(p) },
  { risk: "high", test: (p) => /\.githooks\//.test(p) },
  { risk: "high", test: (p) => /^package\.json$/.test(p) },
  { risk: "medium", test: (p) => /^src\/modules\/(?!auth\/)/.test(p) },
  { risk: "medium", test: (p) => /^db\//.test(p) },
  { risk: "medium", test: (p) => /^scripts\//.test(p) },
  { risk: "low", test: (p) => /^(docs\/|.*\.md$)/.test(p) },
];

function classify(path) {
  for (const rule of RULES) {
    if (rule.test(path)) return rule.risk;
  }
  return "medium";
}

const paths = gitChangedPaths();
let maxRisk = "low";
const riskOrder = { low: 0, medium: 1, high: 2 };

for (const p of paths) {
  const r = classify(p);
  if (riskOrder[r] > riskOrder[maxRisk]) maxRisk = r;
}

console.log(`risk: ${maxRisk}`);
if (paths.length > 0) {
  console.log(`changed: ${paths.length} path(s)`);
  for (const p of paths) console.log(`  [${classify(p)}] ${p}`);
} else {
  console.log("changed: none (clean or single-commit)");
}
