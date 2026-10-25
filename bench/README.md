# bench

Measures whether Delta's Sentinel skill finds real drift between a spec and its code better than a plain prompt does. Development tooling for Delta itself: it is never copied into a project that adopts Delta. The full design and rationale live in `evals.plan.md`.

Everything runs on Bun, with no `node:` imports and no dependencies. Runs use the Claude Code subscription (an `ANTHROPIC_API_KEY` in the environment is deliberately removed for the agent).

## How it works

Each run takes a healthy fixture repo (`repos/shop`), applies one mutation from `cases/`, seals the result in a single git commit, and asks the agent to check one slice. Two modes are compared: `plain` (generic prompt, no skill) and `skill` (the current `.claude/skills/sentinel`). The agent ends its answer with a JSON block of verdicts; `tools/grade.ts` compares it to the truth in code, with no judge model.

## Commands

```
bun test bench/tools                          unit tests for the grader
bun run bench/tools/check.ts                  gate: proves every case is sound (no agent involved)
bun run bench/tools/main.ts --dry             prepare runs and print prompts, without calling the agent
bun run bench/tools/main.ts [options]         run the benchmark
bun run bench/tools/sums.ts <batch>           rebuild a batch's report.md
```

`bun test` at the repo root also runs the fixture's own suite; use `cd bench/repos/shop && bun test` for that one alone.

Options for `main.ts`:

| Option | Meaning |
|---|---|
| `--cases a,b` | only these case ids (default: every non-hold-out case) |
| `--modes a,b` | modes to compare (default `plain,skill`) |
| `--reps N` | repetitions per case and mode (default 3) |
| `--batch NAME` | batch folder under `stats/`; reuse a name to resume |
| `--final` | include the hold-out cases (05, 07, 11); use only for the final batch |
| `--model ID` | model (default `claude-sonnet-5-5`; one model per batch) |
| `--dry` | prepare and print, never call the agent |
| `--keep` | keep the temporary working copies |

## Results

Each batch writes `stats/<batch>/` with `meta.json`, `runs.jsonl`, `errors.jsonl`, `raw/` and `report.md`. Only `report.md` is committed; the rest is git-ignored.

Runs are resumable: finished runs are skipped, errored runs are retried. The runner stops after two consecutive errors (usually a subscription limit); re-run the same command with the same `--batch` to continue.

## Rules of the road

- **Hold-out cases are never run in a tuning batch.** `main.ts` refuses them without `--final`, so Sentinel and the prompts are not tuned against them.
- **Run `check.ts` after touching any case, patch, spec or fixture file.** An unsound case gives a wrong truth and a meaningless result.
- **Changing the model, the fixture or a prompt starts a new batch.** Results from different setups are not comparable.
- Read the "runs worth reading by hand" section of a report and open a few transcripts in `raw/`: the grader can be wrong about the experiment even when it is right about the code.
