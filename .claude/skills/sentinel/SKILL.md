---
name: sentinel
description: Use to verify that a slice's <slice>.spec.md invariants and <slice>.design.md claims still hold against the real implementation — a check, not a rewrite. Trigger when the user asks to verify/audit/check a slice, wants to know if a spec or design is still accurate, or has just finished implementation work on a slice that has a spec or design — e.g. "¿sigue vigente este spec?", "¿esta pantalla sigue cumpliendo su design.md?", "¿la implementación sigue cumpliendo los invariantes de X?", "check this slice for drift". Defaults to a narrow scope (the named slice, or the slice(s) touched by the current diff/recent changes) and never sweeps every spec in the project unasked — token cost for real verification is real, and this skill is designed around bounding it. Reports findings; never edits spec.md, design.md, or source itself — corrections go through `archivist` (and, for a genuine spec/design-vs-code conflict, the same reconciliation discipline `specifier` uses).
---

# Sentinel

`archivist` writes. `specifier` interviews to produce what `archivist` writes. `sentinel` checks whether what's already written is still true — its reason to change is "does the implementation still match its contract," never "is the artifact shaped correctly" or "did we ask the right questions." Same portability rule as its siblings: nothing below assumes a specific stack, and it doesn't check anything outside `<slice>.spec.md` / `<slice>.design.md` — no doc/rule drift-checking, that's a different, broader job this skill deliberately doesn't take on. A spec's or design's optional Context section is out of scope too, deliberately: it's free-form theory, not a set of falsifiable claims, so there's nothing for this skill to check it against.

Real verification means reading real code, not just re-reading a spec — that's inherently more expensive than writing one. The steps below are built around keeping that cost bounded on purpose: narrow scope by default (Step 0), targeted checks before full-file reads (Step 2), and pushing the expensive reading into disposable subagent contexts instead of the main session (Step 2). Skipping these isn't a shortcut, it's the trade-off this skill exists to avoid making.

---

## Step 0 — Scope before anything else

This is the main cost lever — get it narrow before doing any reading.

- **If the user names a slice**, scope is every file that slice name resolves to across the whole repo: `<slice>.spec.md` and `<slice>.design.md` wherever each exists. Nothing wider than that slice. Check the root `deltas/` first — that's where a slice lives by default — and, only in a repo genuinely split into multiple services, every service's own `deltas/` too, not just one. In that multi-service case, a slice's files don't have to share a directory — `design.md` lives under whichever service renders the UI, and each service can even carry its own `<slice>.spec.md` scoped to the logic it owns there (e.g. a backend's session/data contract and a frontend's own navigation-logic contract, same slice name, deliberately different content) — treat those as siblings to check together, not duplicates to collapse into one.
- **If the user doesn't name one** ("check what I just did"), use `git status`/`git diff` to find changed files, map them back to the slice(s) they belong to, and confirm with the user if more than one slice plausibly matches rather than guessing silently.
- **If the user explicitly asks for everything** ("revisa todos los specs"), honor it — but first correlate by slice name across every `deltas/` (root, and every service's own if the repo is genuinely multi-service), so a slice with siblings in two locations is checked and reported as one slice, not two unrelated hits. Say up front that this dispatches one check per slice (not per file) and will cost proportionally, and offer to narrow first. Don't refuse, just don't default to it.
- **If a slice has no `<slice>.spec.md`**, there's nothing to verify against — say so and stop; suggest `specifier`/`archivist` if the user wants one written. Don't invent invariants to check from the implementation alone. Same rule for `<slice>.design.md` if the user specifically asked about the slice's UI and it has none.
- **If a slice has a `spec.md` but no `design.md`**, that's normal — not every slice has a UI-facing surface — just verify what exists and note design verification doesn't apply rather than treating it as a gap.

## Step 1 — Read the contract

Read the scoped `<slice>.spec.md` (and `<slice>.design.md`, if in scope) in full — they're short and bounded, this is the cheap part. Pull out:

- Every line in `spec.md`'s **Invariants** (and any hard guarantee stated in its **Contract**) as one falsifiable claim to check.
- Every line in `design.md`'s **States**, **Interactions**, and **Consumes** (and any hard guarantee stated in **Layout**) as one falsifiable claim to check — e.g. "the empty state shows X," "this action requires confirmation before firing," "this field is sourced from `spec.md`'s Y."

A claim that isn't falsifiable by reading code/behavior isn't something this skill can check — flag it back to the user as unverifiable-as-written rather than forcing a verdict.

## Step 2 — Verify against the real implementation, cheaply

For each claim:

- **Try a targeted check first.** Grep for the specific mechanism the claim names (a check, a guard, a config value) rather than reading whole files end to end — most invariants are mechanically checkable once you know what pattern to look for.
- **Delegate the expensive reading, don't inline it.** If confirming a claim needs broader reading/judgment (not just a grep hit), dispatch it to a subagent instead of pulling that source into the main session's context:
  - If a matching `.claude/agents/*` subagent already exists for this kind of check (e.g. a spec-invariant-checker scoped to this slice), use it — that's exactly what it's for.
  - Otherwise use the `Agent` tool per slice (or per cluster of related claims), so the reading happens in a disposable context and only the verdict comes back. Run independent slices' checks in parallel rather than sequentially.
  - If this kind of check is going to recur for a slice, say so and suggest the user have `archivist` create a dedicated subagent for it (via its own `agents.template.md`) instead of re-paying full inline cost every time.
- **Never guess a verdict.** If what was actually read doesn't clearly confirm or contradict a claim, the verdict is `Unverified`, not a best guess in either direction.

### Escalating to execution

A targeted grep confirms a mechanism is *present*; it can't confirm a mechanism *behaves* a certain way under real input. Most invariants are about presence and the grep above settles them — but a claim stated as behavior ("always validates X", "requires auth", "returns 404 if not found") isn't actually checked by finding the word "validate" in the file. Reporting a Holds on that from grep alone would be trading real confirmation for something that only looks like it.

**When it escalates.** All three have to hold — this is an escalation from the targeted grep above, never a first resort:

- The claim is behavioral/testable, not just the presence of a mechanism.
- The targeted grep was already tried for this specific claim and was inconclusive — it neither confirmed nor contradicted it.
- An automated test already exists that covers this behavior. If none exists, the verdict stays `Unverified` with the note "no executable check available for this claim" — sentinel never writes a new test just to make a claim checkable; that would be inventing verification, not confirming it.

**How it escalates.** Dispatch a subagent — reuse a matching one from `.claude/agents/` if one already covers this kind of check — scoped to that single claim, never to the whole slice, with tools: Read, Glob, Grep, Bash.

- Its system prompt explicitly forbids editing, installing, or writing any file. Bash is only for running the test command that already exists in the project — the same one CI or the project's manifest uses — and reading its result, never for changing anything.
- It returns only the command it ran, the pass/fail result, and the relevant fragment of the output — not the full log; that's what keeps this cheaper than pulling the source into the main session directly.
- If the run fails for a reason unrelated to the claim (a missing env var, broken local setup), the verdict is `Unverified` with that reason. A run that didn't actually exercise what it was meant to test can't be forced into a `Holds` or `Violated` — that would report confidence the check never earned.

**Cost — this is the exception, not the default.** It only triggers when the conditions above are met; it never replaces the targeted grep as the first move, never runs the full test suite (only the specific test or test file tied to the escalated claim), and never changes Step 0's scope — still bounded to the slice(s) already in scope, never a reason to sweep wider on its own.

**When scope spans more than one slice** (a named multi-slice check or a full sweep), do one additional pass after verifying each slice's own claims: compare the Invariants (and, for `design.md`, Interactions/Consumes) already read across those slices — free, since the content is already loaded, no extra reads — for a direct contradiction between two slices' own stated claims on the same domain object. Report only an actual stated contradiction, never an inferred or weak overlap; that becomes a `Conflict` verdict in Step 3.

## Step 3 — Report, don't fix

Report as a compact table, not prose per finding — a verbose report defeats a cost-conscious skill:

| Slice | Source | Claim | Verdict | Evidence |
|---|---|---|---|---|
| `<slice>` | spec / design | `<the claim, restated short>` | Holds / Violated / Unverified / Gone / Conflict | `<what was actually checked — the pattern found or not found>` |

**Evidence must name its own strength.** State plainly whether a verdict came from Step 2's targeted grep or from an escalated execution run — for an execution-backed verdict, name the command and its pass/fail result, not just "test passed." A reader can't judge how much a `Holds` is worth without knowing which kind of check produced it.

- **Holds** — confirmed true against the real implementation.
- **Violated** — confirmed false; the implementation exists and contradicts the claim.
- **Unverified** — couldn't be confirmed either way from what was read; say why.
- **Gone** — the implementation the claim described no longer exists at all (the slice was intentionally retired). Distinct from `Violated`: there's nothing left to contradict, and it isn't a bug to fix.
- **Conflict** — only possible when scope spans more than one slice: two slices' own claims directly contradict each other on the same domain object. The `Slice` column names both slices for this row instead of one.

Lead with anything `Violated`, `Gone`, or `Conflict` — a clean bill of health is the least interesting part of the report.

## Step 4 — Hand off, don't resolve silently

This skill's job ends at the report. What happens next depends on which side is wrong, and that's the user's call, not this skill's to assume:

- **If the code is wrong** (implementation drifted from a still-desired invariant or design claim), that's a bug — say so plainly, don't touch the code.
- **If the spec/design is wrong** (the invariant or design claim no longer applies), hand off to `archivist` to correct the file in place — same as `specifier`'s Step 3, don't pick a side quietly, surface it and let the user confirm which one wins before anything gets edited. If the drift is a `design.md` claim restating something that actually belongs in `spec.md` (or vice versa), flag that shape problem too, not just the factual one.
- **If the verdict is `Gone`** (the implementation is confirmed intentionally removed), hand off to `archivist` to delete the slice's `spec.md`/`design.md` outright — not correct them, not archive them. This system keeps no history layer; git history is already the record of what the contract used to say.
- **If the verdict is `Conflict`**, neither slice is automatically wrong — flag it for the user to decide whether the two slices should be reconciled (two slices whose contracts keep contradicting or changing together are often one bounded context, not two) or whether the overlap is intentional and one claim just needs its scope language tightened.

---

## Non-goals

- Doesn't check `.claude/docs/*` or `.claude/rules/*` for drift — scope is a slice's spec and design only. A broader drift-checker is a different skill, not a reason to stretch this one.
- Doesn't check a spec's or design's optional Context section — it's free-form theory, not a falsifiable contract, so there's nothing to verify it against.
- Doesn't write or edit `spec.md`, `design.md`, or source code — verification only, always handed off.
- Doesn't write or modify tests to make a claim checkable — an escalated execution check only ever runs a test that already exists; if none exists, the claim stays `Unverified`, not a reason to author one.
- Doesn't run the full test suite, even when escalating to execution — only the specific test or test file tied to the claim being escalated. Never a default, always tied to a claim the targeted grep already failed to settle.
- Doesn't sweep every slice in the project by default — see Step 0.
- Doesn't force a Holds/Violated verdict when the evidence genuinely doesn't settle it — `Unverified` is a legitimate, honest outcome, not a failure to try harder.
- Doesn't hunt for cross-slice conflicts when scope is a single slice — the `Conflict` check only runs as a free byproduct of a multi-slice sweep already in scope, never triggers pulling in extra slices or extra reads on its own.