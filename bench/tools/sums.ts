// Aggregates a batch's runs.jsonl into report.md.
// Usage: bun run bench/tools/sums.ts <batch-folder-name>

import type { Usage } from "./exec.ts";
import type { GradedClaim } from "./grade.ts";
import { BENCH } from "./shared.ts";

export type RunRecord = {
  caseId: string;
  category: string;
  holdout: boolean;
  mode: string;
  rep: number;
  slice: string;
  finishedAt: string;
  durationMs: number;
  turns: number;
  usage: Usage;
  subagents: number;
  toolCalls: number;
  permissionDenials: number;
  skillInvoked: boolean;
  formatFailure: boolean;
  clean: boolean;
  driftClaims: number;
  caught: number;
  stableClaims: number;
  falseAlarms: number;
  unverified: number;
  claims: GradedClaim[];
  costUsdNotional: number;
};

export type BatchMeta = {
  batch: string;
  startedAt: string;
  model: string;
  claudeVersion: string;
  bunVersion: string;
  harnessCommit: string;
  sentinelSkillDirty: boolean;
  final: boolean;
  reps: number;
};

const sum = (xs: number[]) => xs.reduce((a, b) => a + b, 0);
const frac = (a: number, b: number) => (b === 0 ? "n/a" : `${a}/${b} (${Math.round((100 * a) / b)}%)`);
const total = (u: Usage) => u.input + u.cacheCreation + u.cacheRead + u.output;

function median(xs: number[]): number {
  if (xs.length === 0) return 0;
  const s = [...xs].sort((a, b) => a - b);
  const mid = Math.floor(s.length / 2);
  return s.length % 2 ? s[mid]! : (s[mid - 1]! + s[mid]!) / 2;
}
const k = (n: number) => (n >= 1000 ? `${(n / 1000).toFixed(1)}k` : String(Math.round(n)));

function row(runs: RunRecord[]): string[] {
  const claims = runs.flatMap((r) => r.claims);
  return [
    String(runs.length),
    frac(runs.filter((r) => r.clean).length, runs.length),
    frac(sum(runs.map((r) => r.caught)), sum(runs.map((r) => r.driftClaims))),
    frac(sum(runs.map((r) => r.falseAlarms)), sum(runs.map((r) => r.stableClaims))),
    frac(claims.filter((c) => c.correct).length, claims.length),
    frac(sum(runs.map((r) => r.unverified)), claims.length),
    frac(runs.filter((r) => r.formatFailure).length, runs.length),
    frac(runs.filter((r) => r.skillInvoked).length, runs.length),
    k(median(runs.map((r) => r.usage.output))),
    k(median(runs.map((r) => total(r.usage)))),
    String(median(runs.map((r) => r.turns))),
    String(Math.round(median(runs.map((r) => r.durationMs)) / 1000)),
    String(median(runs.map((r) => r.subagents))),
  ];
}

const HEAD = [
  "Runs", "Clean runs", "Catch rate", "False alarms", "Claim accuracy", "Unverified",
  "Format failures", "Skill invoked", "Median output tok", "Median total tok", "Median turns", "Median s", "Median subagents",
];

function table(first: string, groups: Map<string, RunRecord[]>): string {
  const lines = [`| ${first} | ${HEAD.join(" | ")} |`, `|${Array(HEAD.length + 1).fill("---").join("|")}|`];
  for (const [name, runs] of groups) lines.push(`| ${name} | ${row(runs).join(" | ")} |`);
  return lines.join("\n");
}

function groupBy(runs: RunRecord[], key: (r: RunRecord) => string): Map<string, RunRecord[]> {
  const out = new Map<string, RunRecord[]>();
  for (const r of runs) out.set(key(r), [...(out.get(key(r)) ?? []), r]);
  return new Map([...out.entries()].sort(([a], [b]) => a.localeCompare(b)));
}

export function buildReport(meta: BatchMeta, runs: RunRecord[]): string {
  const modes = [...new Set(runs.map((r) => r.mode))].sort();
  const caseIds = [...new Set(runs.map((r) => r.caseId))].sort();
  const out: string[] = [];

  out.push(`# Sentinel bench report — ${meta.batch}`, "");
  out.push("| | |", "|---|---|");
  out.push(`| Model | \`${meta.model}\` |`);
  out.push(`| Claude Code | ${meta.claudeVersion} |`);
  out.push(`| Bun | ${meta.bunVersion} |`);
  out.push(`| Harness commit | \`${meta.harnessCommit}\`${meta.sentinelSkillDirty ? " (sentinel skill had uncommitted changes)" : ""} |`);
  out.push(`| Started | ${meta.startedAt} |`);
  out.push(`| Repetitions per case and mode | ${meta.reps} |`);
  out.push(`| Hold-out cases | ${meta.final ? "included (final batch)" : "excluded (tuning batch)"} |`);
  out.push(`| Runs | ${runs.length} |`, "");
  out.push(
    "Reading guide: counts are shown next to percentages because there are few runs; a gap of under about 15 points is noise. " +
      "*Catch rate* is drift claims reported correctly out of drift claims; *False alarms* is stable claims flagged `Violated`/`Gone` out of stable claims. " +
      "Tokens are the sum of input, cache creation, cache read and output; dollar cost is not reported because the runs use a subscription.",
    "",
  );

  out.push("## By mode", "", table("Mode", groupBy(runs, (r) => r.mode)), "");
  out.push("## By category and mode", "", table("Category / mode", groupBy(runs, (r) => `${r.category} / ${r.mode}`)), "");

  out.push("## Clean runs by case", "");
  out.push(`| Case | ${modes.join(" | ")} |`, `|---|${modes.map(() => "---").join("|")}|`);
  for (const id of caseIds) {
    const cells = modes.map((m) => {
      const rs = runs.filter((r) => r.caseId === id && r.mode === m);
      return rs.length === 0 ? "-" : `${rs.filter((r) => r.clean).length}/${rs.length}`;
    });
    const mark = runs.find((r) => r.caseId === id)?.holdout ? " (hold-out)" : "";
    out.push(`| ${id}${mark} | ${cells.join(" | ")} |`);
  }
  out.push("");

  const missed = runs.filter((r) => r.driftClaims > r.caught || r.falseAlarms > 0 || r.formatFailure);
  if (missed.length > 0) {
    out.push("## Runs worth reading by hand", "");
    for (const r of missed) {
      const why = [
        r.formatFailure ? "format failure" : "",
        r.driftClaims > r.caught ? "missed drift" : "",
        r.falseAlarms > 0 ? `${r.falseAlarms} false alarm(s)` : "",
      ].filter(Boolean).join(", ");
      out.push(`- \`${r.caseId}\` / ${r.mode} / rep ${r.rep}: ${why}`);
    }
    out.push("");
  }
  return out.join("\n");
}

export async function readRuns(batchDir: string): Promise<RunRecord[]> {
  const file = Bun.file(`${batchDir}/runs.jsonl`);
  if (!(await file.exists())) return [];
  return (await file.text()).split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l) as RunRecord);
}

if (import.meta.main) {
  const batch = Bun.argv[2];
  if (!batch) {
    console.error("usage: bun run bench/tools/sums.ts <batch-folder-name>");
    process.exit(2);
  }
  const dir = `${BENCH}/stats/${batch}`;
  const meta = (await Bun.file(`${dir}/meta.json`).json()) as BatchMeta;
  await Bun.write(`${dir}/report.md`, buildReport(meta, await readRuns(dir)));
  console.log(`wrote ${dir}/report.md`);
}
