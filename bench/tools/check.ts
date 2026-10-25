// Soundness gate for the benchmark: no agent is involved.
// For every case it applies the mutation to a fresh copy of the repo, runs the
// fixture's tests, and verifies that the case's metadata and truth are coherent.
// Usage: bun run bench/tools/check.ts [case-id ...]

type Verdict = "Holds" | "Violated" | "Unverified" | "Gone" | "Conflict";
type Claim = { spec: string; n: number; ok: Verdict[] };
type Base = { repo: string; claims: Claim[] };
type Case = {
  id: string;
  repo: string;
  category: string;
  slice: string;
  holdout?: boolean;
  patch: string | null;
  expect: { tests: "pass" | "fail"; failing?: string[] };
  overrides: Claim[];
};

const BENCH = `${import.meta.dir}/..`;
const TEMP = (process.env.TEMP ?? process.env.TMPDIR ?? "/tmp").replaceAll("\\", "/");
const CATEGORIES = ["ctrl", "pres", "beht", "bent", "gone", "beni", "spec"];
const VERDICTS: Verdict[] = ["Holds", "Violated", "Unverified", "Gone", "Conflict"];

function run(cmd: string[], cwd: string): { code: number; out: string } {
  const r = Bun.spawnSync(cmd, { cwd, stdin: "ignore" });
  return { code: r.exitCode ?? 1, out: r.stdout.toString() + r.stderr.toString() };
}

async function freshCopy(repo: string): Promise<string> {
  const dir = `${TEMP}/bench-check-${crypto.randomUUID().slice(0, 8)}`;
  const from = `${BENCH}/repos/${repo}`;
  // Bun's shell `cp -r` panics on Windows (Bun 1.3.14), so copy file by file.
  for await (const rel of new Bun.Glob("**/*").scan({ cwd: from, dot: true, onlyFiles: true })) {
    await Bun.write(`${dir}/${rel.replaceAll("\\", "/")}`, Bun.file(`${from}/${rel}`));
  }
  return dir;
}

function runTests(dir: string): { passed: boolean; failing: string[] } {
  const r = run([process.execPath, "test"], dir);
  const failing = [...r.out.matchAll(/^\(fail\) (.+?)(?: \[[\d.]+ms\])?$/gm)].map((m) => m[1]!);
  return { passed: r.code === 0, failing };
}

async function countInvariants(dir: string, spec: string): Promise<number> {
  const file = Bun.file(`${dir}/${spec}`);
  if (!(await file.exists())) return 0;
  const text = (await file.text()).replace(/\r\n/g, "\n");
  const section = text.split(/^## Invariants\s*$/m)[1]?.split(/^## /m)[0] ?? "";
  return [...section.matchAll(/^\d+\. /gm)].length;
}

async function dirHasFiles(dir: string, path: string): Promise<boolean> {
  for await (const _ of new Bun.Glob(`${path}/**`).scan({ cwd: dir })) return true;
  return false;
}

function touchedFiles(patch: string): string[] {
  return [...patch.matchAll(/^diff --git a\/(\S+) b\/\S+$/gm)].map((m) => m[1]!);
}

const only = Bun.argv.slice(2);
const caseIds: string[] = [];
for await (const meta of new Bun.Glob("*/case.json").scan({ cwd: `${BENCH}/cases` })) {
  caseIds.push(meta.replaceAll("\\", "/").split("/")[0]!);
}
const selected = caseIds.filter((id) => only.length === 0 || only.includes(id)).sort();

const results: { id: string; problems: string[]; note: string }[] = [];
const cleanup: string[] = [];

// Baseline: the untouched repo must be healthy and match the base truth.
const baseProblems: string[] = [];
const base = (await Bun.file(`${BENCH}/cases/shop.base.json`).json()) as Base;
{
  const dir = await freshCopy(base.repo);
  cleanup.push(dir);
  const t = runTests(dir);
  if (!t.passed) baseProblems.push(`healthy repo fails its own tests: ${t.failing.join("; ")}`);
  for (const spec of new Set(base.claims.map((c) => c.spec))) {
    const have = base.claims.filter((c) => c.spec === spec).map((c) => c.n).sort((a, b) => a - b);
    const want = Array.from({ length: await countInvariants(dir, spec) }, (_, i) => i + 1);
    if (JSON.stringify(have) !== JSON.stringify(want)) {
      baseProblems.push(`${spec}: base truth covers [${have}] but the spec has invariants [${want}]`);
    }
  }
  for (const c of base.claims) {
    if (c.ok.length === 0 || c.ok.some((v) => !VERDICTS.includes(v))) {
      baseProblems.push(`${c.spec} #${c.n}: invalid ok set`);
    }
  }
}
results.push({ id: "(baseline)", problems: baseProblems, note: `${base.claims.length} claims` });

for (const id of selected) {
  const problems: string[] = [];
  const c = (await Bun.file(`${BENCH}/cases/${id}/case.json`).json()) as Case;

  // Metadata coherence.
  if (c.id !== id) problems.push(`id "${c.id}" does not match folder`);
  if (!CATEGORIES.includes(c.category)) problems.push(`unknown category "${c.category}"`);
  if (!id.includes(`-${c.category}-`)) problems.push(`folder name does not carry category "${c.category}"`);
  if (c.repo !== base.repo) problems.push(`repo "${c.repo}" has no base truth`);

  const patchFile = c.patch ? `${BENCH}/cases/${id}/${c.patch}` : null;
  if (c.category === "ctrl" && c.patch !== null) problems.push("a control must not have a patch");
  if (c.category !== "ctrl" && c.patch === null) problems.push("a non-control case needs a patch");

  const specPath = `deltas/${c.slice}.spec.md`;
  const dir = await freshCopy(c.repo);
  cleanup.push(dir);

  // The patch must stay inside the repo and touch the right kind of file.
  if (patchFile) {
    if (!(await Bun.file(patchFile).exists())) {
      problems.push(`patch file missing: ${c.patch}`);
    } else {
      const files = touchedFiles(await Bun.file(patchFile).text());
      if (c.category === "spec" && !files.every((f) => f.startsWith("deltas/"))) {
        problems.push("a spec-side case may only touch deltas/");
      }
      if (c.category !== "spec" && files.some((f) => f.startsWith("deltas/"))) {
        problems.push("a code-side case must not touch deltas/");
      }
      const applied = run(["git", "apply", "--whitespace=nowarn", patchFile], dir);
      if (applied.code !== 0) problems.push(`patch does not apply: ${applied.out.trim().split("\n")[0]}`);
    }
  }

  // Test outcome must match the declared expectation.
  const t = runTests(dir);
  const note = t.passed ? "tests pass" : `tests fail (${t.failing.length})`;
  if (c.expect.tests === "pass" && !t.passed) problems.push(`expected tests to pass, failing: ${t.failing.join("; ")}`);
  if (c.expect.tests === "fail") {
    if (t.passed) problems.push("expected tests to fail, but they pass");
    for (const want of c.expect.failing ?? []) {
      if (!t.failing.some((f) => f.includes(want))) problems.push(`expected failing test not seen: "${want}"`);
    }
    const extra = t.failing.filter((f) => !(c.expect.failing ?? []).some((w) => f.includes(w)));
    if (extra.length > 0) problems.push(`unexpected failing tests: ${extra.join("; ")}`);
  }
  if (c.category !== "beht" && c.expect.tests !== "pass") problems.push(`category "${c.category}" must leave the suite green`);
  if (c.category === "beht" && c.expect.tests !== "fail") problems.push('category "beht" must break a test');

  // Truth coherence against the mutated repo.
  const count = await countInvariants(dir, specPath);
  if (count === 0) problems.push(`${specPath} has no numbered invariants`);
  for (const o of c.overrides) {
    if (o.spec !== specPath) problems.push(`override for ${o.spec} is outside the scoped slice`);
    if (o.n < 1 || o.n > count) problems.push(`override #${o.n} is outside 1..${count}`);
    if (o.ok.length === 0 || o.ok.some((v) => !VERDICTS.includes(v))) problems.push(`override #${o.n}: invalid ok set`);
  }
  if (new Set(c.overrides.map((o) => o.n)).size !== c.overrides.length) problems.push("duplicate override");

  if (c.category === "ctrl" || c.category === "beni") {
    if (c.overrides.length > 0) problems.push(`category "${c.category}" must not override any claim`);
  } else if (c.category === "gone") {
    const all = Array.from({ length: count }, (_, i) => i + 1);
    const goneAll = all.every((n) => c.overrides.some((o) => o.n === n && o.ok.length === 1 && o.ok[0] === "Gone"));
    if (!goneAll) problems.push("a retired slice must mark every claim as Gone");
    if (await dirHasFiles(dir, `source/${c.slice}`)) problems.push(`source/${c.slice} still exists in a "gone" case`);
  } else {
    if (c.overrides.length === 0) problems.push("a mutation case needs at least one override");
    if (!c.overrides.some((o) => o.ok.length === 1 && o.ok[0] === "Violated")) {
      problems.push("a mutation case needs a claim whose only acceptable verdict is Violated");
    }
  }

  results.push({ id, problems, note });
}

for (const dir of cleanup) await Bun.$`rm -rf ${dir}`.quiet();

let failed = 0;
for (const r of results) {
  const mark = r.problems.length === 0 ? "ok  " : "FAIL";
  if (r.problems.length > 0) failed++;
  console.log(`${mark} ${r.id.padEnd(22)} ${r.note}`);
  for (const p of r.problems) console.log(`       - ${p}`);
}
console.log(failed === 0 ? `\nall ${results.length} checks passed` : `\n${failed} of ${results.length} checks failed`);
process.exit(failed === 0 ? 0 : 1);
