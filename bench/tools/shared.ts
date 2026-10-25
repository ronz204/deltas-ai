// Types and small helpers shared by the bench tools. Bun APIs only.

export type Verdict = "Holds" | "Violated" | "Unverified" | "Gone" | "Conflict";
export type Claim = { spec: string; n: number; ok: Verdict[] };
export type Base = { repo: string; claims: Claim[] };
export type Case = {
  id: string;
  repo: string;
  category: string;
  slice: string;
  holdout?: boolean;
  patch: string | null;
  expect: { tests: "pass" | "fail"; failing?: string[] };
  overrides: Claim[];
};
export type Mode = {
  id: string;
  // Skills copied from this repository's .claude/skills into the working copy.
  skills: string[];
  // Value of --allowedTools, space separated.
  tools: string;
};

export const BENCH = import.meta.dir.replaceAll("\\", "/").replace(/\/tools$/, "");
export const ROOT = BENCH.replace(/\/bench$/, "");
export const TEMP = (process.env.TEMP ?? process.env.TMPDIR ?? "/tmp").replaceAll("\\", "/");
export const MODEL = "claude-sonnet-5-5";
export const VERDICTS: Verdict[] = ["Holds", "Violated", "Unverified", "Gone", "Conflict"];

export function run(cmd: string[], cwd: string): { code: number; out: string } {
  const r = Bun.spawnSync(cmd, { cwd, stdin: "ignore" });
  return { code: r.exitCode ?? 1, out: r.stdout.toString() + r.stderr.toString() };
}

// Bun's shell `cp -r` panics on Windows (Bun 1.3.14), so copy file by file.
export async function copyTree(from: string, to: string): Promise<void> {
  for await (const rel of new Bun.Glob("**/*").scan({ cwd: from, dot: true, onlyFiles: true })) {
    await Bun.write(`${to}/${rel.replaceAll("\\", "/")}`, Bun.file(`${from}/${rel}`));
  }
}

export async function hasFiles(dir: string): Promise<boolean> {
  try {
    for await (const _ of new Bun.Glob("**/*").scan({ cwd: dir, dot: true, onlyFiles: true })) return true;
  } catch {
    // A missing folder simply has no files.
  }
  return false;
}

export async function removeTree(dir: string): Promise<void> {
  await Bun.$`rm -rf ${dir}`.quiet();
}

export function tempDir(prefix: string): string {
  return `${TEMP}/${prefix}-${crypto.randomUUID().slice(0, 8)}`;
}

export async function loadBase(repo: string): Promise<Base> {
  return (await Bun.file(`${BENCH}/cases/${repo}.base.json`).json()) as Base;
}

export async function loadCases(only: string[] = []): Promise<Case[]> {
  const ids: string[] = [];
  for await (const meta of new Bun.Glob("*/case.json").scan({ cwd: `${BENCH}/cases` })) {
    ids.push(meta.replaceAll("\\", "/").split("/")[0]!);
  }
  const cases: Case[] = [];
  for (const id of ids.sort()) {
    if (only.length > 0 && !only.includes(id)) continue;
    cases.push((await Bun.file(`${BENCH}/cases/${id}/case.json`).json()) as Case);
  }
  return cases;
}

export async function loadMode(id: string): Promise<Mode> {
  return (await Bun.file(`${BENCH}/modes/${id}/mode.json`).json()) as Mode;
}

// The claims that apply to a case: the base truth for the scoped spec, with the
// case's overrides replacing the matching entries.
export function truthFor(base: Base, c: Case): Claim[] {
  const spec = `deltas/${c.slice}.spec.md`;
  return base.claims
    .filter((claim) => claim.spec === spec)
    .map((claim) => c.overrides.find((o) => o.spec === claim.spec && o.n === claim.n) ?? claim);
}

export function render(template: string, vars: Record<string, string>): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key: string) => vars[key] ?? "");
}
