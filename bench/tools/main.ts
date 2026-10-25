// Runs the benchmark: cases x modes x repetitions, resumable.
//
//   bun run bench/tools/main.ts [options]
//
//   --cases a,b      only these case ids (default: every non-hold-out case)
//   --modes a,b      modes to compare (default: plain,skill)
//   --reps N         repetitions per case and mode (default: 3)
//   --batch NAME     batch folder under bench/stats; reuse a name to resume
//   --final          include the hold-out cases (only for the final batch)
//   --model ID       model to run (default: the fixed bench model)
//   --dry            prepare each run and print the prompt, without calling the agent
//   --keep           keep the working copies instead of deleting them

import { execute } from "./exec.ts";
import { grade } from "./grade.ts";
import { prepare } from "./prep.ts";
import {
  BENCH, MODEL, ROOT, loadBase, loadCases, loadMode, removeTree, render, run, truthFor,
  type Case, type Mode,
} from "./shared.ts";
import { buildReport, readRuns, type BatchMeta, type RunRecord } from "./sums.ts";

function parseArgs(argv: string[]) {
  const out = { cases: [] as string[], modes: ["plain", "skill"], reps: 3, batch: "", final: false, model: MODEL, dry: false, keep: false };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]!;
    const next = () => argv[++i] ?? "";
    if (a === "--cases") out.cases = next().split(",").filter(Boolean);
    else if (a === "--modes") out.modes = next().split(",").filter(Boolean);
    else if (a === "--reps") out.reps = Number(next());
    else if (a === "--batch") out.batch = next();
    else if (a === "--model") out.model = next();
    else if (a === "--final") out.final = true;
    else if (a === "--dry") out.dry = true;
    else if (a === "--keep") out.keep = true;
    else throw new Error(`unknown option: ${a}`);
  }
  if (!Number.isInteger(out.reps) || out.reps < 1) throw new Error("--reps must be a positive integer");
  return out;
}

const args = parseArgs(Bun.argv.slice(2));

// Hold-out cases are reserved for the final batch so tuning never sees them.
const all = await loadCases(args.cases);
const requestedHoldout = all.filter((c) => c.holdout && args.cases.includes(c.id));
if (requestedHoldout.length > 0 && !args.final) {
  throw new Error(`${requestedHoldout.map((c) => c.id).join(", ")} is a hold-out case; pass --final to run it`);
}
const cases = all.filter((c) => args.final || !c.holdout);
if (cases.length === 0) throw new Error("no cases selected");
const modes: Mode[] = await Promise.all(args.modes.map(loadMode));

const batch =
  args.batch ||
  new Date().toISOString().slice(0, 16).replace("T", "-").replace(":", "");
const dir = `${BENCH}/stats/${batch}`;
const metaFile = Bun.file(`${dir}/meta.json`);
const meta: BatchMeta = (await metaFile.exists())
  ? ((await metaFile.json()) as BatchMeta)
  : {
      batch,
      startedAt: new Date().toISOString(),
      model: args.model,
      claudeVersion: run(["claude", "--version"], ROOT).out.trim(),
      bunVersion: Bun.version,
      harnessCommit: run(["git", "rev-parse", "--short", "HEAD"], ROOT).out.trim(),
      sentinelSkillDirty: run(["git", "status", "--porcelain", "--", ".claude/skills/sentinel"], ROOT).out.trim() !== "",
      final: args.final,
      reps: args.reps,
    };
if (meta.model !== args.model) throw new Error(`batch ${batch} was started with ${meta.model}; a batch has one model`);
if (!args.dry) await Bun.write(`${dir}/meta.json`, JSON.stringify(meta, null, 2) + "\n");

const runs: RunRecord[] = args.dry ? [] : await readRuns(dir);
const done = new Set(runs.map((r) => `${r.caseId}|${r.mode}|${r.rep}`));

// Interleave: repetition-major, and alternate which mode goes first so neither
// systematically benefits from running earlier (cache warmth, rate limits).
type Job = { c: Case; mode: Mode; rep: number };
const jobs: Job[] = [];
for (let rep = 1; rep <= args.reps; rep++) {
  cases.forEach((c, i) => {
    const order = (rep + i) % 2 === 0 ? modes : [...modes].reverse();
    for (const mode of order) if (!done.has(`${c.id}|${mode.id}|${rep}`)) jobs.push({ c, mode, rep });
  });
}
console.log(`batch ${batch}: ${jobs.length} runs to do (${done.size} already done), model ${args.model}`);

const suffix = await Bun.file(`${BENCH}/modes/suffix.md`).text();
const baseCache = new Map<string, Awaited<ReturnType<typeof loadBase>>>();
let consecutiveErrors = 0;

for (const [index, { c, mode, rep }] of jobs.entries()) {
  const label = `[${index + 1}/${jobs.length}] ${c.id} / ${mode.id} / rep ${rep}`;
  const vars = { slice: c.slice };
  const template = await Bun.file(`${BENCH}/modes/${mode.id}/prompt.md`).text();
  const prompt = `${render(template, vars).trim()}\n\n${render(suffix, vars).trim()}`;

  const workdir = await prepare(c, mode);
  try {
    if (args.dry) {
      console.log(`${label}\n  workdir: ${workdir}\n  tools: ${mode.tools}\n  prompt: ${prompt.replaceAll("\n", "\n          ")}`);
      continue;
    }

    const started = Date.now();
    const exec = await execute(workdir, prompt, mode.tools, args.model);
    await Bun.write(`${dir}/raw/${c.id}.${mode.id}.${rep}.jsonl`, exec.raw);

    if (exec.isError || exec.exitCode !== 0 || exec.timedOut || exec.resultText === "") {
      const problem = { caseId: c.id, mode: mode.id, rep, at: new Date().toISOString(), exitCode: exec.exitCode, timedOut: exec.timedOut, stderr: exec.stderr.slice(0, 500), result: exec.resultText.slice(0, 500) };
      const errFile = Bun.file(`${dir}/errors.jsonl`);
      await Bun.write(errFile, ((await errFile.exists()) ? await errFile.text() : "") + JSON.stringify(problem) + "\n");
      console.log(`${label}: ERROR (exit ${exec.exitCode}${exec.timedOut ? ", timed out" : ""}) ${exec.stderr.slice(0, 120).trim()}`);
      if (++consecutiveErrors >= 2) {
        console.log(`stopping after ${consecutiveErrors} consecutive errors (a subscription limit is the usual cause). Resume with: --batch ${batch}`);
        break;
      }
      continue;
    }
    consecutiveErrors = 0;

    if (!baseCache.has(c.repo)) baseCache.set(c.repo, await loadBase(c.repo));
    const g = grade(exec.resultText, truthFor(baseCache.get(c.repo)!, c));
    const record: RunRecord = {
      caseId: c.id, category: c.category, holdout: c.holdout === true, mode: mode.id, rep, slice: c.slice,
      finishedAt: new Date().toISOString(), durationMs: exec.durationMs || Date.now() - started,
      turns: exec.turns, usage: exec.usage, subagents: exec.subagents, toolCalls: exec.toolCalls,
      permissionDenials: exec.permissionDenials, skillInvoked: exec.skillInvoked,
      formatFailure: g.formatFailure, clean: g.clean,
      driftClaims: g.driftClaims, caught: g.caught, stableClaims: g.stableClaims, falseAlarms: g.falseAlarms,
      unverified: g.unverified, claims: g.claims, costUsdNotional: exec.costUsd,
    };
    runs.push(record);
    await Bun.write(`${dir}/runs.jsonl`, runs.map((r) => JSON.stringify(r)).join("\n") + "\n");
    console.log(`${label}: ${g.formatFailure ? "FORMAT FAILURE" : g.clean ? "clean" : "not clean"}; caught ${g.caught}/${g.driftClaims}, false alarms ${g.falseAlarms}; ${Math.round(record.durationMs / 1000)}s, ${exec.turns} turns, skill ${exec.skillInvoked ? "yes" : "no"}`);
  } finally {
    if (!args.keep) await removeTree(workdir);
    else console.log(`  kept ${workdir}`);
  }
}

if (!args.dry) {
  await Bun.write(`${dir}/report.md`, buildReport(meta, runs));
  console.log(`report: ${dir}/report.md`);
}
