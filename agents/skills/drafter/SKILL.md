---
name: drafter
description: >-
  Use to draft or revise the contract of a slice, i.e. to write its <slice>.spec.md. It investigates the real code, interviews the user about intent, scope, invariants and acceptance, then writes the claims with `delta spec write` as a draft for a person to approve. comper hands the work to it whenever behavior is new or changes, and for a small amendment (a limit, a value, one claim's wording) of a slice that is already built. Trigger when the user wants to "spec out" a slice, write or harden a spec, or asks what a spec is missing, e.g. "let's spec out X", "write the spec for X", "this spec needs hardening", /drafter <slice>. Never approves and never edits files under deltas/ directly. Not for a trivial single-file slice with no invariant worth protecting.
argument-hint: <slice>
---

# Drafter

Code shows what a slice currently does. It doesn't show why the slice exists, what is deliberately out of scope, what must never break, or what is intentionally left undecided. Those only come from asking. This skill produces that missing half and writes it down as a contract: claims with stable IDs, each with a way to verify it. Its reason to change is "did we ask the right questions and find the right conflicts", never "does the code satisfy the contract" (that is `builder`'s) or "is the contract still true" (that is `auditor`'s).

Drafter owns the span of the cycle up to `draft`. A person owns the next step: `delta approve`. Nothing below assumes a specific stack.

The slice is `$ARGUMENTS`; if it is empty, ask which slice. Interview the user in their language, but write the spec in English.

---

## Step 0 — Read what already exists

- Run `delta status` to see the slices and their state.
- If `<slice>` has a spec (`deltas/specs/<slice>.spec.md`), read it in full. Every question below must be about a genuine gap or an update, not something already answered. Note its state: rewriting an approved or built contract sends it back to `draft`, and a person must approve it again.
- If it has none, confirm the slice's name and boundary with the user first. The file is created in Step 4, once there is content for it.

## Amend: a small change to a slice that is already built

When the slice is `built` or `verified` and the request changes one or a few claims (a limit, a value, the wording of one invariant) without changing the slice's intent or scope, don't run the whole interview. Do this instead of Steps 1 to 3:

- Read the claims that change and the code that governs them. Ask at most one question, only if the new value or wording is not already clear.
- Write the whole spec with `delta spec write` (Step 4), with only those claims changed. IDs stay stable, and the frontmatter is `status: draft`.
- Show the person **only** the before and after of each changed claim, and say whether its tag still names a test. Ask them to approve it.

If the request adds claims, or changes the slice's intent, scope or contract, it is not an amendment: go on with Step 1.

An approved amendment doesn't need a rebuild when its tests still hold: `comper` runs `delta check <slice>` next, and `builder` is called only if a claim fails.

## Step 1 — Investigate the real implementation

Find and read the source that implements the slice: grep by the feature's name, walk its folder, read whole files rather than excerpts. Don't infer behavior from a filename or a doc's description.

While reading, check the implementation against any project doc that describes the slice and any project rule that should apply to it. Note every mismatch: a doc that says one thing while the code does another is exactly the gap a contract exists to close, so it is surfaced (Step 3), never quietly resolved in either direction.

## Step 2 — Ask what the code alone can't answer

For each point, confirm it from what Step 1 found or ask the user. Don't guess.

- **Intent**: why does the slice exist, in a sentence or two of domain language?
- **Scope and non-goals**: what is explicitly not this slice's job, especially anything a future contributor could assume belongs here?
- **Invariants**: what must hold however the implementation changes? This is the highest-value part, because it is what gets checked later. Phrase each as one EARS claim whose response can be observed (`When <trigger>, <subject> <response>`; also `While`, `Where`, `If … then`, or always/never). If the user's wording is vague or joins two outcomes, ask what triggers it and what would be seen, and propose the rewrite.
- **Acceptance criteria**: each as `Given <context>, when <action>, then <outcome>`, only for what the invariants don't already cover.
- **Deferred / open questions**: what is intentionally undecided, and what event would decide it.
- **How each claim is verified**, because it decides the tag:
  - `test: <command>`: a command that exits 0 when the claim holds. Name the file, package or folder it runs and don't filter by test name (`go test ./billing`, not `-run X`), so a filter that matches nothing can't pass silently. Find the existing test and confirm it exists. A claim with no test yet is not left out: write it as `none`.
  - `manual: <step>`: only a person can check it.
  - `none`: nothing verifies it yet. It is valid, stays visible as debt in `delta status`, and is listed under Deferred as a testing gap.

Ask in a short, focused batch (`AskUserQuestion` when the choices are concrete) rather than the whole list at once. A spec built from five rushed answers is worse than one built from two considered ones; continue next turn if needed.

## Step 3 — Reconcile conflicts explicitly

Any mismatch from Step 1, and any place where an answer contradicts a doc, a rule or the slice's current spec, is surfaced and resolved with the user before it goes anywhere. Never silently pick the code's version or the doc's version. Note which side won and why; the resolution is folded into the spec. There is no separate decision log.

## Step 4 — Write the draft with the CLI

The knowledge base is not edited directly: a hook blocks `Write` and `Edit` on `deltas/`. The CLI is the only way in, and it validates before it saves.

1. If the slice has no spec, run `delta spec new <slice>`. It creates the file from the template as a `draft`. Read it to see the sections.
2. Compose the whole spec and send it on stdin with the Bash tool and a heredoc: `delta spec write <slice> <<'EOF'` … `EOF`. Use Bash even where PowerShell is the default shell: a piped PowerShell here-string can't be matched by a permission rule, so it would stop for approval every time. The frontmatter keeps `format: 1`, `slice: <bare name>` and `status: draft` (a new spec can only be a draft; the other states are moved by commands).
3. Each claim is a header line `- INV-01 [tag]` or `- ACC-01 [tag]`, and under it, indented, its one sentence. IDs are stable: never renumber an existing claim, and don't reuse the ID of one you removed.
4. If it prints diagnostics, nothing was written. Fix what they say (they carry file, line and code) and write again. When it succeeds, run `delta spec lint <slice>`.

Then stop. Tell the user the slice is a `draft`, summarize the claims and which are `none` or `manual`, and ask them to review it. **Don't run `delta approve`**: the sign-off is the person's, and Claude Code will ask them anyway. The next steps are `delta approve <slice>` and then `/builder <slice>`.

If the resulting spec implies real implementation work, that is `builder`'s job, not a deliverable of this skill.

---

## Non-goals

- Doesn't approve, build or verify. It ends at `draft`.
- Doesn't write into `deltas/` any other way than `delta spec write`, even when the answer seems obvious enough to write directly.
- Doesn't re-ask what the current spec already settles.
- Doesn't force every field to a firm answer. "Intentionally undecided until X" is a complete answer for Deferred; a premature decision makes the spec lie about its own certainty.
- Doesn't spec a slice that is a single trivial file with no invariant worth protecting. The ceremony pays for itself against real complexity, not by default.
- Doesn't write docs, rules or the project's `CLAUDE.md`. Those are edited by a person.
