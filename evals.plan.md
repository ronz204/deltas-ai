# Evals plan — `bench/` (Sentinel by mutation)

A planning document. **Nothing in here is implemented yet.** It exists so the folder structure, the experiment design, and the open risks can be reviewed before any code is written. Once approved, each phase below is implemented and committed separately.

Origin: `NEXT.md` section 4.A ("Measure before building") and the decision to start with the Sentinel family because it has exact ground truth.

---

## 1. Goal and the one question it answers

> **Does the Sentinel skill find real drift between a spec and its code better than an agent given no skill and a plain prompt — and at what token cost?**

Not in scope for this first round: Surveyor quality, implementation-with-context tasks, trimmed-KB experiments. Those are later families on the same skeleton.

Why Sentinel first:

- **Exact ground truth.** We break an invariant on purpose, so we know the right answer. No judge model, no opinions.
- **It is Delta's differentiator.** Drift detection is the niche nobody else covers (see `NEXT.md` section 3).
- **It is cheap to make deterministic.** A verdict table is machine-checkable.

## 2. Vocabulary

| Term | Meaning here |
|---|---|
| **Repo** | A small TypeScript project used as the environment. Healthy by default: tests pass, every spec invariant holds. |
| **Case** | One scenario: a repo plus one *mutation* (a patch that breaks, or deliberately does not break, something) plus the truth about what the right verdict is. |
| **Mode** | How the agent is equipped: which files are overlaid on the repo and which prompt it receives. This is the thing being compared. |
| **Run** | One execution of one case under one mode. Every run is repeated several times because the model is not deterministic. |
| **Stats** | Results of runs: raw transcripts, one line of metrics per run, an aggregated report. |
| **Mutation** | A patch applied to the healthy repo before the run. |
| **Truth** | For each numbered invariant of each spec, the set of verdicts that count as correct. |

## 3. Folder structure

Every direct subfolder of `bench/` has a five-letter name, so the tree reads evenly.

```
bench/
├── README.md                     how to run, how to read results
├── .gitattributes                forces LF so patches apply on any checkout
├── .gitignore                    ignores everything under stats/<batch>/ except report.md
├── repos/                        ENVIRONMENTS (copied, never edited in place)
│   └── shop/                     fixture #1: a tiny TypeScript shop backend
│       ├── package.json          no dependencies; `bun test` runs bun's built-in test runner
│       ├── tsconfig.json         path aliases @pricing, @stock, @checkout for cross-slice imports
│       ├── CLAUDE.md             minimal router, as a real Delta project would have
│       ├── source/
│       │   ├── pricing/          slice: money, discounts, tax
│       │   ├── stock/            slice: inventory and reservations
│       │   └── checkout/         slice: orchestrates pricing + stock + payment
│       ├── tests/                tests for behavior-style invariants (some invariants intentionally have none)
│       └── deltas/
│           ├── pricing.spec.md   numbered invariants, healthy
│           ├── stock.spec.md
│           └── checkout.spec.md
├── cases/                        SCENARIOS (mutation + truth)
│   ├── shop.base.json            truth for the healthy repo: acceptable verdicts per invariant
│   ├── 01-ctrl-checkout/
│   │   └── case.json             no mutation file: a control
│   ├── 03-pres-stockevent/
│   │   ├── case.json             metadata + truth overrides
│   │   └── patch.diff            the mutation
│   └── ...                       12 cases in total (section 6)
├── modes/                        THE VARIABLE BEING COMPARED
│   ├── suffix.md                 the fixed output-format instruction appended to every prompt
│   ├── plain/
│   │   ├── mode.json             skills to overlay (none) and the tools allowed
│   │   └── prompt.md             generic "verify this spec against the code" prompt
│   └── skill/
│       ├── mode.json             skills: ["sentinel"], copied from .claude/skills at run time
│       └── prompt.md             natural-language "check the <slice> slice for drift"
├── tools/                        RUNNER (TypeScript, zero dependencies, run with bun)
│   ├── shared.ts                 types and small helpers used by every tool
│   ├── check.ts                  validates that every case is sound (phase 2 gate)
│   ├── prep.ts                   builds a sealed working copy for one run
│   ├── exec.ts                   launches the agent headless and extracts the metrics
│   ├── grade.ts                  compares the agent's verdicts against truth
│   ├── grade.test.ts             unit tests for the grader (run with `bun test bench/tools`)
│   ├── sums.ts                   aggregates runs into the markdown report
│   └── main.ts                   entry point: loops cases x modes x repetitions
└── stats/                        RESULTS
    ├── .gitkeep
    └── 2026-11-02-1530/          one folder per batch
        ├── meta.json             model, versions, harness commit (local)
        ├── runs.jsonl            one line of metrics per run; resumable (local)
        ├── errors.jsonl          runs that errored, retried on resume (local)
        ├── raw/                  full transcripts (local)
        └── report.md             aggregated tables; the only committed file
```

Naming rules:

- Top-level subfolders: `repos`, `cases`, `modes`, `tools`, `stats` (all five letters).
- Case folders: `NN-kind-short` (e.g. `06-beht-rounding`), where `kind` is one of the four-letter category codes in section 6.
- Nothing under `bench/` is ever copied into a user's project. A project adopting Delta copies `.claude/` (and `.cursor/`) only; `bench/` is development tooling for Delta itself.

### Why `repos` and `cases` are separate

A repo is the expensive thing to build and must be reusable; a case is a cheap patch plus a few lines of truth. Many cases hang off one repo. (This is the correction to the earlier "one folder per project" idea.)

## 4. How one run works

For every combination of case, mode, and repetition:

1. **Prepare.** Copy `repos/<repo>` to a fresh temporary directory *outside* the Delta repository (so no parent `CLAUDE.md` can leak in).
2. **Mutate.** Apply `cases/<id>/patch.diff` if it exists.
3. **Overlay.** Copy `modes/<mode>/files/` on top (for `skill`, this includes the current `.claude/skills/sentinel/` copied from this repository, so every run tests the skill as it is today).
4. **Seal.** `git init` and make a **single commit** that includes the mutation. This is deliberate: if the mutation sat in the working tree, the agent could find it with `git diff` and the test would measure nothing.
5. **Ask.** Launch the agent headless in that directory with the mode's prompt plus a fixed output-format suffix (section 5).
6. **Capture.** Save the transcript, token usage, turn count, wall time, whether the Sentinel skill was invoked, and how many subagents ran.
7. **Grade.** Parse the verdict block and compare to truth (section 7).
8. **Record.** Append one line to `runs.jsonl`.

Truth files and patches are never inside the working copy. The agent sees only the sealed repo.

## 5. The two modes (first round)

| | `plain` | `skill` |
|---|---|---|
| Sentinel skill present | no | yes (current version) |
| Prompt style | "Read `deltas/<slice>.spec.md` and check each invariant against the code. Report the result for each one." | "Check the `<slice>` slice for drift." (natural trigger; the skill must be chosen and followed by the agent) |
| Specs and code | identical | identical |
| Output suffix | identical | identical |

Fixed output suffix appended to both prompts, so grading is deterministic:

> End your answer with a fenced `json` block of the form `{"verdicts":[{"spec":"deltas/<slice>.spec.md","n":3,"verdict":"Holds"}]}`, one entry for every item under each scoped spec's **Invariants** heading, where `n` is the item's 1-based position and `verdict` is one of `Holds`, `Violated`, `Unverified`, `Gone`, `Conflict`.

Because this suffix changes how Sentinel reports, it is listed as a threat to validity (section 10). Invariants in the fixture specs are numbered lists so `n` is unambiguous.

Possible later modes (not in round one): `forced` (explicit `/sentinel` invocation, to separate "did it trigger" from "does it work"), `lite` (the deterministic-checks variant from `NEXT.md` B2), a trimmed-instructions variant.

## 6. The fixture and the cases

### 6.1 Fixture `shop`

A deliberately small TypeScript backend with three slices. Invariant types are mixed on purpose so Sentinel's different paths are exercised: *presence* (a grep settles it), *behavior with an existing test* (Sentinel may escalate to run that test), and *behavior with no test* (the honest answer is `Unverified` or a clear read of the code).

Sketch of the specs. The final wording lives in `bench/repos/shop/deltas/*.spec.md` (written in phase 1) and is authoritative; where this sketch differs (for example stock 1 is "per order and SKU", and unit prices must be non-negative so that "a total is never negative" truly holds), the files win:

| Slice | # | Invariant (sketch) | Type |
|---|---|---|---|
| pricing | 1 | Money is integer cents; no floating point in price arithmetic | presence |
| pricing | 2 | Percentage discounts round half up to the nearest cent | behavior, test |
| pricing | 3 | A total is never negative; a discount is capped at the subtotal | behavior, test |
| pricing | 4 | Tax is computed on the discounted subtotal | behavior, test |
| pricing | 5 | The pricing module is pure: no I/O, no clock reads | presence |
| pricing | 6 | Mixing currencies throws | behavior, test |
| stock | 1 | Reservations are idempotent per `orderId` | behavior, test |
| stock | 2 | Available = on hand − reserved and is never negative | behavior, test |
| stock | 3 | Every stock mutation emits a `StockChanged` event | presence |
| stock | 4 | Reservations expire after 15 minutes (`RESERVATION_TTL_MS`) | behavior, **no test** |
| stock | 5 | Stock never imports from checkout (dependency direction) | presence |
| checkout | 1 | Checkout of an empty cart throws `EmptyCartError` | behavior, test |
| checkout | 2 | Stock is reserved before charging and released if payment fails | behavior, test |
| checkout | 3 | Payment is attempted at most once per `orderId` | behavior, test |
| checkout | 4 | Order totals come only from the pricing module (no inline arithmetic) | presence |
| checkout | 5 | An order is persisted only after successful payment | behavior, test |

Technical constraints:

- **Zero dependencies.** The project's runtime is Bun: tests use `bun:test` and run with `bun test`, TypeScript needs no build step. The fixture is written as plain, portable TypeScript (no `enum`, no `namespace`, no parameter properties) so it would also read naturally in other runtimes.
- **Healthy baseline is enforced.** `bun test` must pass on the untouched repo, and `cases/shop.base.json` records that every invariant is in an acceptable state.
- The fixture's `deltas/*.spec.md` are **test data hand-written to control the invariants exactly**, so they are exempt from the "route through `archivist`" rule that governs this project's own knowledge base. The plan's phase 1 should state this exemption where it is created so it is not mistaken for drift.

### 6.2 The twelve cases

| # | Id | Category | Mutation | Expected verdict for the touched claim |
|---|---|---|---|---|
| 01 | `01-ctrl-checkout` | `ctrl` control | none | all `Holds` (or `Unverified` where no test exists) |
| 02 | `02-ctrl-pricing` | `ctrl` control | none | same |
| 03 | `03-pres-stockevent` | `pres` presence | remove one `StockChanged` emit | stock 3 `Violated` |
| 04 | `04-pres-inlinemath` | `pres` presence | checkout computes a total inline instead of via pricing | checkout 4 `Violated` |
| 05 | `05-pres-floatmoney` | `pres` presence | introduce a float multiplication in pricing | pricing 1 `Violated` |
| 06 | `06-beht-rounding` | `beht` behavior, test | half-up rounding becomes floor | pricing 2 `Violated` |
| 07 | `07-beht-retrypay` | `beht` behavior, test | payment is retried a second time | checkout 3 `Violated` |
| 08 | `08-bent-ttl` | `bent` behavior, no test | TTL changed from 15 to 60 minutes | stock 4 `Violated` (only; `Unverified` counts as a miss) |
| 09 | `09-gone-stock` | `gone` retired | the stock implementation and its tests are deleted, spec left behind | every stock claim `Gone` |
| 10 | `10-beni-rename` | `beni` benign | behavior-preserving rename and file move in pricing | all `Holds` (tests the false-alarm rate) |
| 11 | `11-beni-extract` | `beni` benign | extract a helper in checkout, same behavior | all `Holds` |
| 12 | `12-spec-rounding` | `spec` spec-side | the spec is edited to claim half-even rounding; the code is unchanged | pricing 2 `Violated` |

Mix: 2 controls, 3 presence, 2 behavior-with-test, 1 behavior-without-test, 1 retired, 2 benign, 1 spec-side. Controls and benign cases (4 of 12) exist to catch an agent that cries wolf, which matters as much as catching real drift.

A `Conflict` case (two slices contradicting each other) is deferred to round two because it needs a multi-slice scope.

### 6.3 Case and truth file shapes

`cases/shop.base.json` records what is acceptable on the healthy repo:

```json
{
  "repo": "shop",
  "claims": [
    { "spec": "deltas/pricing.spec.md", "n": 1, "ok": ["Holds"] },
    { "spec": "deltas/stock.spec.md",   "n": 4, "ok": ["Holds", "Unverified"] }
  ]
}
```

`cases/06-beht-rounding/case.json` records only what changes:

```json
{
  "id": "06-beht-rounding",
  "repo": "shop",
  "category": "beht",
  "slice": "pricing",
  "patch": "patch.diff",
  "expect": { "tests": "fail", "failing": ["rounds half up to the nearest cent"] },
  "overrides": [
    { "spec": "deltas/pricing.spec.md", "n": 2, "ok": ["Violated"] }
  ]
}
```

`slice` is what the prompt names, which keeps Sentinel's scope narrow (its own Step 0). `ok` is a set, not a single value, because honest uncertainty (`Unverified`) is a legitimate answer for some claims. `expect` declares what the fixture's own test suite must do after the patch (`pass`, or `fail` naming the failing tests); it is only used by `tools/check.ts` to prove the case is sound, and is never shown to the agent. `holdout: true` marks the reserved cases (05, 07, 11). `patch` is `null` for controls.

Notes from building the cases:

- Case 09 removes the stock slice *coherently* (its source, its tests, its wiring in checkout, and the checkout tests that depended on it), so the repo stays green and the retirement looks intentional. The stock spec is left behind on purpose; every stock claim is `Gone`.
- Case 10 renames helpers that are not named in any spec's Contract (`lineTotal` to `priceLine`, `min` to `smaller`), so the rename cannot be mistaken for contract drift.
- Case 05 uses float arithmetic that still rounds half up. This was checked exhaustively (cents 0 to 200000, percent 0 to 100: zero differences from the integer version), so claim 2 genuinely still holds and only claim 1 is violated.
- `bench/.gitattributes` forces LF line endings so patches apply identically on any checkout.

### 6.4 Tooling rules

All code under `bench/tools/` is Bun-only: Bun APIs (`Bun.file`, `Bun.write`, `Bun.Glob`, `Bun.spawnSync`, `Bun.$`, `import.meta.dir`) and no imports starting with `node:`. Known Bun 1.3.14 limitation on Windows: the shell builtin `cp -r` panics, so directories are copied file by file with `Bun.Glob` and `Bun.write`.

## 7. Grading and metrics

Grading is pure code, with no model judging.

Per run:

- **Parse.** Extract the final JSON block. If it is missing or malformed, the run is a *format failure* (counted, not hidden).
- **Claim check.** For each claim in the scoped spec, is the reported verdict in that claim's `ok` set?
- **Clean run.** True if every claim in the scoped spec is within its `ok` set.

Aggregated, per mode and per category:

| Metric | Definition | What it tells us |
|---|---|---|
| **Catch rate** | Of claims whose `ok` is `Violated`/`Gone`, share reported correctly | Does it find real drift? |
| **False-alarm rate** | Of claims whose `ok` excludes `Violated`/`Gone`, share reported as `Violated`/`Gone` | Does it cry wolf? |
| **Claim accuracy** | Share of all claims reported within `ok` | Overall correctness |
| **Clean-run rate** | Share of runs that are fully correct | The strict, easy-to-explain number |
| **Unverified rate** | Share of claims answered `Unverified` | Honesty versus dodging |
| **Format failures** | Runs with no parseable block | Prompt/skill robustness |
| **Skill invoked** | In `skill` mode, share of runs where the skill was actually used | Triggering reliability |
| **Cost** | Input and output tokens, turns, wall time, subagent count (median and spread) | The price of the result |

Because the plan is run on a subscription, **dollar cost is not the metric**. Tokens, turns, and time are.

Reporting rule: always show counts next to percentages (e.g. `8/9`), because there are few runs.

## 8. Run matrix and budget

- Round one: **12 cases × 2 modes × 3 repetitions = 72 runs.**
- Order is interleaved (not all `plain` then all `skill`) so that drift over time or rate-limiting does not bias one mode.
- Runs are **resumable**: `runs.jsonl` is read first and finished combinations are skipped, so a rate-limit stop on the subscription costs nothing.
- Token use per run is unknown until the pilot (phase 4). A `skill` run may spawn subagents and can be several times more expensive than a `plain` run; the pilot measures this before the full batch is committed to.

## 9. Interpreting results

Decide the reading rules *before* seeing data:

| Outcome | Reading | Next step |
|---|---|---|
| `skill` clearly better on catch rate, equal or better false alarms | Sentinel adds value | Invest in `NEXT.md` B1/B2 |
| `skill` and `plain` about equal | The skill's structure matters less than expected; a plain prompt does the same job | Question the skill's complexity; prioritize the deterministic checks (B2/B4) |
| `skill` better only on behavior-with-test | The escalation to run tests is the valuable part | Strengthen and cheapen that path |
| `skill` catches more but false alarms rise | It over-reports | Tighten the verdict rules |
| `skill` rarely invoked | The description does not trigger | Fix the trigger before judging the content |
| `skill` equal but far more tokens | Cost without benefit | Reduce its procedure |

With 72 runs this is a signal detector, not a statistical proof. A gap of less than about 15 percentage points should be treated as noise; the pilot's variance sets the real threshold.

## 10. Risks and threats to validity

| Risk | Why it matters | Mitigation |
|---|---|---|
| **Environment leakage** | The headless agent may still load the user's global skills, settings, or memory, contaminating `plain` mode or helping it. | Phase 0 spike must prove isolation (how to detect: ask the plain agent to list its available skills). The difficulty: isolating the config directory must not lose the subscription login. If isolation cannot be achieved, the limitation is written into the report instead of hidden. |
| **Mutation visible in git** | `git diff` or history would reveal it. | Single sealed commit (section 4, step 4). |
| **Truth or patch visible to the agent** | Would void the experiment. | Both live outside the working copy. |
| **Output suffix changes Sentinel** | The skill's own report format is a table; we ask for JSON as well. | The suffix is identical across modes and appended after the report; noted in every report. |
| **Author bias** | One person designs the fixture and the mutations. | Keep two or three cases unseen by the skill's author, and add a second fixture in round two, ideally a real open source repo. |
| **Small fixture** | A tiny repo flatters any approach. | Treat results as relative (`plain` vs `skill`), not absolute; add a bigger repo later. |
| **Nondeterminism** | One run proves nothing. | Three repetitions minimum, spread reported. |
| **Overfitting** | Tuning Sentinel until these 12 cases pass. | Hold-out cases `05`, `07`, `11` (section 12) are never inspected while tuning. |
| **Bun version drift** | Behavior of `bun test` can change between versions. | Record the Bun version in each batch's report; round one was built on Bun 1.3.14. |
| **Cases that are not actually sound** | A "behavior" mutation that does not fail its test makes truth wrong. | `tools/check.ts` is a mandatory gate (phase 2): it applies each patch, runs the fixture tests, and asserts the expected pass or fail. |
| **Windows paths and shells** | Development happens on Windows. | Use path APIs (`node:path` works under Bun), `git apply` for patches, and no shell-specific scripts. |

## 11. Implementation phases

Each phase ends in a checked state and its own commit, per `commit-conventions`. Suggested commit types are in parentheses.

**Phase 0 — Spike (answers the unknowns; throwaway findings recorded here, no bench code yet)**
- Run the agent headless once and confirm the exact command-line flags for non-interactive use, structured output, tool restriction, and model selection.
- Confirm what the output reports: tokens, turns, tool calls, how a skill invocation shows up.
- Test environment isolation (section 10, first risk).
- Confirm the TypeScript runtime on this machine (Bun; see section 14).
- Output: a short findings block appended to this file.

**Phase 1 — Fixture** (`test(bench)`)
- Build `bench/repos/shop`, its three specs with numbered invariants, its tests, and `shop.base.json`.
- Gate: `npm test` green on the untouched repo; a manual read confirms each invariant genuinely holds.

**Phase 2 — Cases and the soundness checker** (`test(bench)`)
- Write the 12 mutations and their `case.json` files, plus `tools/check.ts`.
- Gate: `check.ts` passes for all 12 (patches apply, behavior-with-test cases really fail the right test, benign cases still pass everything, controls untouched).

**Phase 3 — Runner, grader, report** (`feat(bench)`)
- `prep`, `exec`, `grade`, `sums`, `main`, plus `bench/README.md` and `.gitignore`.
- Gate: a dry run on one control case produces a well-formed `runs.jsonl` line and report.

**Phase 4 — Pilot** (no commit until reviewed)
- One case in each category, both modes, three repetitions.
- Look at variance, token use, and read several raw transcripts by hand to confirm grading matches reality. Adjust prompts or truth if the experiment itself is flawed.

**Phase 5 — Full batch and write-up** (`docs(bench)`)
- Run the full matrix, commit `report.md`, and record conclusions and the decision (section 9) in `NEXT.md`.

## 12. Decisions

Settled:

1. **Model: Sonnet 5.5** (`claude-sonnet-5-5`), fixed for every run in every mode. Changing the model starts a new, non-comparable batch.
2. **Only `report.md` is committed.** `raw/` transcripts stay git-ignored.
3. **Case 08 demands `Violated`.** The TTL constant is readable in code, so reading it against the spec settles the claim; answering `Unverified` counts as a miss. (The healthy-repo baseline for stock 4 still accepts `Holds` or `Unverified`, since the untouched repo has no test for it.)
4. **Round two uses a new repository**, ideally a real open source project pinned to a commit, chosen after round one's results.
5. **Hold-out cases: `05`, `07`, `11`** (one presence, one behavior-with-test, one benign). While tuning Sentinel or the prompts, their per-case results are not looked at; they are read only in the final batch. The reason: with only 12 cases, tuning against all of them can fit the skill to those cases without improving it in general, and the hold-outs are the check that the improvement is real. Proposed by Claude; change freely.

Still open: none for round one.

## 13. Non-goals

- Not a general agent benchmark; it measures Delta's own skills only.
- No LLM judge anywhere in grading.
- Does not evaluate Surveyor, Specifier, Archivist, or implementation tasks in round one.
- Does not change `.claude/skills/sentinel` or any harness file; the benchmark only reads the current skill.
- Not shipped to projects that adopt Delta.

---

## 14. Phase 0 findings (2026-10-25)

Spike done on Windows, Claude Code 2.1.300, Node v24.11.1, git 2.53.0.

**Runtime: Bun, not Node (correction).** The spike checked Node 24 (native TypeScript works there), but this project works with Bun 1.3.14, so the fixture and the runner use Bun: `bun test` with `bun:test`, `bun run tools/main.ts`. No `package.json` dependency is needed. Imports inside the fixture use the `.ts` extension.

**Headless command shape.** `claude -p "<prompt>" --model claude-sonnet-5-5 --output-format stream-json --verbose --no-session-persistence --setting-sources project --strict-mcp-config --permission-mode dontAsk --allowedTools "<list>"`, run with the sealed working copy as the current directory and stdin redirected from `/dev/null` (otherwise a 3-second "no stdin data" warning appears).

**Isolation (the first risk in section 10), resolved for round one.**

| Flags | What the agent can see |
|---|---|
| none | 41 skills, commands and MCP tools: the account's plugins and connectors leak in |
| `--setting-sources project --strict-mcp-config` | 13 items, all skills built into Claude Code itself (`code-review`, `init`, `loop`...), plus the project's own `.claude/skills` |
| `--safe-mode` | the same 13, but it also hides the project's skill, so it cannot be used for the `skill` mode |
| `--bare` | not usable: it never reads the subscription login |

The login keeps working with `--setting-sources project --strict-mcp-config`. There is no user-level `CLAUDE.md` on this machine; `~/.claude/skills` holds only a `synced` folder, which did not appear in the listing. Decision: both modes use `--setting-sources project --strict-mcp-config`. The 13 built-in skills are present in both modes and are not under test; to keep the comparison symmetric, `plain` mode's allowed tools exclude `Skill`, so it cannot invoke any of them.

**Permissions.** `--permission-mode dontAsk` with an explicit allow list worked: a full Sentinel run finished with no permission denials. Allow list for both modes: `Read Grep Glob Agent Bash(git status:*) Bash(git diff:*) Bash(bun test:*)`, plus `Skill` in `skill` mode only. Nothing that writes files is allowed.

**What the JSON output gives us** (`result` event): `usage` (input, output, cache read and cache creation tokens), `num_turns`, `duration_ms`, `subagent_stats.spawned`, `permission_denials`, `total_cost_usd` (a list-price figure, notional on a subscription), and the final `result` text. With `stream-json`, each tool call is an event, so **skill invocation is detected** by a `tool_use` named `Skill` whose input is `sentinel`.

**Smoke test.** A two-invariant spec over a 6-line module (one rounding bug planted: `Math.floor` where the spec says half up) ran in 7 turns and 20 seconds with no subagents. The skill was invoked, the planted bug was reported as `Violated`, the other invariant as `Holds`, and the closing JSON block was well-formed. So the output suffix in section 5 is parseable in practice.

**Consequences for later phases.**

- Phase 1: imports use the `.ts` extension; tests run with `bun test`.
- Phase 3: `exec.ts` parses `stream-json`; the allow list above is part of `modes/*/mode.json`.
- Cost is small for a trivial slice (about 20 s); the full cost per run is still to be measured in the phase 4 pilot.
- Not yet verified: how much a run costs when Sentinel delegates to subagents or runs a test; whether `dontAsk` blocks the `bun test` escalation path as intended (verify in the pilot).

---

## 15. Phase 3 results (2026-10-25)

Built: `shared.ts`, `prep.ts`, `exec.ts`, `grade.ts` (+ `grade.test.ts`), `sums.ts`, `main.ts`, both modes with `modes/suffix.md`, `bench/README.md`, `bench/.gitignore`. `check.ts` now uses `shared.ts`.

Gate, as run:

- `bun test bench/tools`: 9 grader tests pass (extraction of the last JSON block, catch and false-alarm counting, `Unverified` on a drifted claim counted as a miss, omitted claims, format failures, path spelling, `Gone`).
- `check.ts`: all 13 checks still pass after the refactor.
- A dry run showed the sealed copies: one commit, mutation committed, clean tree, the Sentinel skill present only in the `skill` copy, and no patch or truth file anywhere in the working copy.
- A real smoke run (control case `01-ctrl-checkout`, both modes, one repetition, batch deleted afterwards): both clean; `skill` mode invoked the skill, `plain` did not; about 13 s and 5 to 6 turns each; no permission denials. Token medians for that single control were about 60k (`plain`) and 93k (`skill`) in total, mostly cache reads. One sample says nothing about the difference; it only shows the metrics are captured.

Design decisions made while building:

- The prompt goes to the agent through stdin, so no argument quoting can alter it.
- `ANTHROPIC_API_KEY` is removed from the agent's environment so runs cannot silently bill the API instead of the subscription.
- `main.ts` refuses hold-out cases unless `--final` is passed, which turns the plan's "never inspect hold-outs while tuning" into a mechanism instead of a promise.
- Which mode goes first alternates per case and repetition, so neither benefits from always running earlier.
- Errors are logged to `errors.jsonl` and retried on resume; two consecutive errors stop the batch, since the usual cause is a subscription limit.
- Modes have no `files/` overlay folder yet (the code supports one if a mode ever needs it).
- Known weakness, to confirm in the pilot: `dontAsk` permissions with the `Bash(bun test:*)` allow rule have not yet been exercised by a case that makes Sentinel run a test.

---

## 16. Phase 4 results: the pilot (2026-10-25)

Batch `pilot`: 7 non-hold-out cases (one per category: `01`, `03`, `06`, `08`, `09`, `10`, `12`) x 2 modes x 3 repetitions = 42 runs, `claude-sonnet-5-5`, harness commit `6e94472`, zero errors. Its `report.md` is committed under `bench/stats/pilot/`. The hold-out cases (05, 07, 11) were not run.

### What the numbers say

| | `plain` | `skill` |
|---|---|---|
| Clean runs | 21/21 | 21/21 |
| Catch rate (drift claims found) | 27/27 | 27/27 |
| False alarms | 0/87 | 0/87 |
| Format failures | 0/21 | 0/21 |
| Skill invoked | n/a | 21/21 |
| Median total tokens | 61.7k | 121.3k |
| Median turns | 7 | 11 |
| Median seconds | 12 | 18 |
| Subagents spawned (all runs) | 0 | 0 |

Total tokens summed over all runs: `skill` used 1.68 times what `plain` used. Results were identical across the three repetitions of every case.

### How to read it

1. **The benchmark saturated.** Both modes scored 100% everywhere, so it cannot tell them apart. This is a ceiling effect of a fixture that is too easy, not evidence that the skill is unnecessary: each slice has 3 to 5 source files and 5 to 6 invariants, so any agent can read the whole slice in a few calls and reach the right verdict. Sentinel's machinery (targeted grep before reading, delegating to subagents, escalating to run tests) is designed for slices too big to read whole.
2. **The skill is reliably triggered.** 21/21 invocations from a natural-language prompt, with no instruction to use it.
3. **At equal accuracy the skill costs about 1.7 times the tokens and 1.5 times the time.** Its procedure adds turns that bought nothing on this fixture. By the reading table in section 9 ("equal accuracy, far more tokens") that is tentatively cost without benefit, but only on this easy fixture; it must be re-tested on a harder one before drawing a conclusion.
4. **Grading was validated by hand.** Two full answers (`08-bent-ttl`, both modes) were read: the verdicts match the code and each cites the right line. In the `skill` answer the human table and the JSON block agree.
5. **No subagent was ever spawned**, and no run used `Unverified`. Neither the delegation path nor honest uncertainty was exercised.

### An experiment flaw found and fixed

On Windows the agent runs commands through the `PowerShell` tool, not `Bash`, and the allow list only covered `Bash(...)`. In 4 `skill` runs (`03` reps 2 and 3, `06` reps 1 and 2) the agent tried `bun test ...` and was denied; in `plain` it tried a directory listing and was denied. Answers stayed correct because the agent fell back to reading, but the execution-escalation path of Sentinel was effectively blocked during the pilot. Fix: both `mode.json` files now also allow `PowerShell(bun test:*)`, `PowerShell(git status:*)`, `PowerShell(git diff:*)` and `PowerShell(Select-Object:*)` (the agent pipes test output through it). A probe run of `06` in `skill` mode afterwards executed `bun test` with zero denials. The pilot numbers above were produced before the fix; because every run was 100% correct they are not expected to change, but the pilot report should be read with that in mind.

### Consequence for the plan

Running the full 72-run batch on these same cases would very likely repeat the ceiling and spend subscription tokens for no information. Recommendation: do not run it as designed. Build a harder "round 1b" first (section 17), then run the full batch on that.

## 17. Making the benchmark discriminate (proposal, not started)

Goal: cases where a plain prompt can plausibly fail and the skill's discipline can matter.

1. **A bigger slice.** More files per slice and 10 to 15 invariants per spec, so reading everything is expensive and the targeted-check and delegation steps of Sentinel become relevant.
2. **Subtler mutations.** Off-by-one boundaries (`>=` vs `>`), a default changed in a deep helper, a swapped call order (release before charge), an error type swapped for a sibling, a constant defined in two places, an invariant violated only through a transitive import via a barrel file.
3. **Mutations only execution reveals.** Misleading names or stale comments that suggest the claim holds while a test fails, so the skill's escalation to run an existing test earns its keep.
4. **Several drifts in one slice** (two or three at once) to stress recall, not just single-needle detection.
5. **Decoys.** Scary-looking but behavior-preserving changes (moved code, renamed constants, restructured but equivalent logic) to keep measuring false alarms where they are likelier.
6. **A `Conflict` case** (two slices contradicting each other, multi-slice scope), deferred earlier.
7. **Bring forward round two's real open source repository.** A synthetic fixture the author wrote and mutated is the easiest possible material; a real repo with real specs written for it is the stronger test.

Whichever subset is chosen, `check.ts` must pass for every new case before any agent run, and the hold-out discipline applies to the new set.
