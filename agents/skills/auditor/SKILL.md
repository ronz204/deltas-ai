---
name: auditor
description: >-
  Use to verify a slice against its contract: `delta check` runs the commands of its test claims, and this skill reads the code only for the claims no command covers (manual and none). It reports Holds, Violated or Unverified per claim and never edits anything. comper hands the work to it once a slice is built. Trigger when the user asks to verify, audit or certify a slice, wants to know if a spec is still accurate, or has just finished implementing one, e.g. "is this spec still accurate?", "does X still meet its invariants?", "check this slice for drift", /auditor <slice>. Defaults to one slice and never sweeps every spec unasked.
argument-hint: <slice>
---

# Auditor

`drafter` writes the contract, `builder` builds it, `auditor` checks it. Its reason to change is "does the implementation still match its contract", never "is the contract worded well" or "how do we build it".

The check has two halves, and only the second needs a model:

- **Deterministic: `delta check`.** It runs the command of every `test` claim and judges it from the exit code and from whether a test under the command's paths cites the claim ID. No model reads anything. This half is cheap, repeatable, and settles most claims.
- **Reading: the claims with no command.** A `manual` or `none` claim can only be judged by reading the code. That is the only place this skill spends tokens, and it bounds the cost on purpose: narrow scope, targeted greps before full reads, and subagents for the expensive reading.

Nothing below assumes a specific stack. The slice is `$ARGUMENTS`.

---

## Step 0 — Scope before anything else

This is the main cost lever.

- If the user names a slice, scope is that slice and nothing wider.
- If they don't, use `git status` and `git diff` to find the slices the changes belong to, and confirm with the user if more than one fits instead of guessing.
- If they ask for everything, honor it, but say first that it costs proportionally and offer to narrow. Don't refuse, and don't default to it.
- If the slice has no spec, there is nothing to verify against: say so and suggest `drafter`. Don't invent claims from the code.

## Step 1 — Run the deterministic half

Run `delta spec lint <slice>` (it must be clean), then `delta check <slice>`, and keep its table: one row per claim, with `Holds`, `Violated` or `Unverified` and the reason. Running `delta check` also moves a `built` slice to `verified` when every `test` claim holds; that is expected and is not a write by this skill.

Read the verdicts as `delta check` gives them:

- `Holds` is confirmed: the command exited 0 and a test cites the ID. It does not prove the test ran (a skipped test passes), so mention `delta lint` warnings (`LINT-SKIP`) when there are any.
- `Violated` is a failing command. Report it; the code or the test is wrong.
- `Unverified` with a reason ("no test it runs cites the claim ID", "could not run the command") is a link or setup problem, not a reading job. Report it as drift in the spec's tag or in the environment. Don't substitute a judgment of your own.

## Step 2 — Read only what no command covers

Take the `manual` and `none` claims, and nothing else.

- **Targeted check first.** Grep for the specific mechanism the claim names (a guard, a check, a config value) rather than reading whole files; most claims settle once you know what pattern to look for.
- **Delegate the expensive reading.** If confirming a claim needs broader reading, dispatch it to a subagent (the `Agent` tool, or a matching one under `.claude/agents/`) so the source stays out of the main context and only the verdict comes back. Run independent claims in parallel.
- **Never run what writes.** Reading a claim may involve running an existing read-only command; never one that edits, installs or deletes, and never the whole suite.
- **Never guess.** If what you read doesn't clearly confirm or contradict a claim, the verdict is `Unverified`, not a best guess. A claim that names no observable response can't be checked: send it back to `drafter` as unverifiable as written.
- **A claim that is checked by reading stays `none` or `manual`.** The reading verdict is a report, not a state change: `delta check` still counts the claim as unverified, because nothing runs it.

## Step 3 — Report, don't fix

Report a compact table, one row per claim, leading with anything `Violated`:

| Slice | Claim | Verdict | Checked by | Evidence |
|---|---|---|---|---|
| `<slice>` | `<the claim, short>` | Holds / Violated / Unverified | command / reading | `<the command and its result, or the pattern found or not found>` |

State plainly which kind of check produced each verdict, because a `Holds` from a command is worth more than one from reading. End with one line that counts them, for example `11 claims: 6 by command, 2 by reading, 3 unverified (1 manual, 2 none)`, so the report doubles as a coverage summary.

## Step 4 — Hand off, don't resolve silently

The skill ends at the report. What happens next depends on which side is wrong, and that is the user's call.

- **The code is wrong** (it drifted from a still-wanted claim): say so plainly and don't touch it. If the user wants it fixed, that is `builder`.
- **The spec is wrong** (the claim no longer applies, or its tag is stale): hand it to `drafter`, which rewrites it through `delta spec write` and sends the slice back to `draft` for a new approval. Surface it and let the user say which side wins.
- **A `none` claim that reading shows holds** is a gap in the tests: suggest `builder` write the test and link it with `delta link`, so the next `delta check` can verify it by execution.
- **The implementation is gone** (the slice was retired): say so; removing the spec file is the user's call.
- **Two slices' claims contradict each other** on the same object (only visible when scope spans several slices, as a byproduct of reading them): flag it, don't pick a side.

---

## Non-goals

- Doesn't write or edit specs, tests or source. Verification only, always handed off.
- Doesn't write a test to make a claim checkable.
- Doesn't approve, build or change a slice's state. `delta check` does the one move it is allowed to do.
- Doesn't check docs or rules for drift: its scope is a slice's contract.
- Doesn't sweep every slice by default.
- Doesn't force a `Holds` or `Violated` when the evidence doesn't settle it. `Unverified` is an honest outcome.
