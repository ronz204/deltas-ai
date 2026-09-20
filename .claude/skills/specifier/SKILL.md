---
name: specifier
description: Use before writing or substantially revising a <slice>.spec.md (or <slice>.design.md) under deltas/ at the project root (or, in a repo genuinely split into multiple services, under a service's own deltas/ directory). Gathers the information a robust spec or design actually needs — investigating the real implementation, checking it against existing docs/rules, and asking the user targeted questions about intent, boundaries, invariants, and open decisions, or, for the UI-facing counterpart, layout/states/interactions/what it consumes — then hands that material off to the archivist skill to write into <slice>.spec.md and/or <slice>.design.md. Trigger whenever the user wants to "spec out" a slice, flesh out or harden an existing spec, define how a screen/surface should look and behave, or says things like "vamos a especificar X", "necesito robustecer el spec de Y", "hagamos el spec de este slice", "¿qué le falta a este spec?", "definamos el design de esta pantalla". Never writes the spec or design file directly — that's archivist's job; this skill only produces the grounded material archivist needs.
---

# Specifier

Code shows what a slice currently does. It doesn't show why the slice exists, what's deliberately out of scope, what must never break, or what's intentionally left undecided — those only come from asking. The same gap exists on the presentation side: a mockup shows one snapshot of a screen, not every state it must handle, what happens on each interaction, or what it's allowed to depend on from the slice's logic-layer contract. This skill's entire job is producing that missing half so the spec (and, where relevant, the design) `archivist` writes is grounded and complete, not a paraphrase of the source code or a mockup with the important parts guessed at. Like `archivist`, this skill is meant to be portable across projects — its steps describe a way of investigating and interviewing, not anything specific to one codebase's stack or module names.

Splitting this from `archivist` is a deliberate SRP call: this skill's reason to change is "did we ask the right questions and find the right conflicts" — `archivist`'s reason to change is "does the artifact land in the right shape." Keeping them separate means the writing conventions (templates, phrasing, English-only, no product framing) stay consistent regardless of who's answering questions on a given day.

A **delta spec** is what this skill produces material for: `<slice>.spec.md` describes a *delta* — a bounded change or capability inside a project, specified as a contract for how that artifact/sector of the project should behave — rather than an account of how it got that way. The slice is the unit; the spec is its current, living contract for the logic layer. A slice with a UI-facing surface has a second living contract, `<slice>.design.md`, for the presentation layer — same "current truth, not history" discipline, gathered the same interviewing way, just aimed at layout/states/interactions/consumption instead of data/rules/invariants.

---

## Step 0 — Read what already exists for this slice

Before asking the user anything, find out what's already known:

- Locate `deltas/<slice>.spec.md` at the project root and, if present, `<slice>.design.md` alongside it. Only in a repo genuinely split across multiple services, don't assume every sibling sits next to the file you started from: `design.md` lives under whichever service renders the UI, which can be a different service than the one implementing the logic, and the same slice name can carry its *own* `spec.md` in more than one service (each scoped to what that service owns) rather than a single shared one — check other services' `deltas/` for both before concluding a sibling doesn't exist. If any already have content, read them — every question below should be about a genuine gap or an update, not something already answered.
- If the slice's spec (or design) file doesn't exist yet, don't create it — confirm the slice's name and boundary with the user first (creating the file is `archivist`'s step, once there's real content to put in it).

## Step 1 — Investigate the real implementation

Same discipline `archivist` applies at its own Step 0: find and read the actual source that implements this slice — grep by the slice's feature/path-alias name, walk its folder, read the whole file rather than an excerpt. Don't infer behavior from a filename or a related doc's description.

While reading, actively check the implementation against:
- Any `.claude/docs/*` claim that describes this slice.
- Any `.claude/rules/*` convention that should apply to it (e.g. a naming convention, an access-control pattern, a data-handling rule — whatever this project's actual rules cover).

Note every mismatch found — a doc that says one thing while the code does another is exactly the kind of gap a robust spec exists to close, not something to quietly resolve in either direction.

Also watch for a pattern that clearly extends beyond this one slice — not "how this capability behaves" but "how code gets written in this project" (a naming convention, a required check, a structural pattern repeated regardless of slice). That's a rule candidate, not spec content — see `archivist`'s "Spotting a rule candidate" check (cross-cutting, not already tool-enforced, applies every time, would be silently violated otherwise). Don't fold it into this slice's `Invariants`; carry it into Step 3 instead.

## Step 2 — Ask what code (or a mockup) alone can't answer

For each of the following, either confirm it directly from what Step 1 found, or ask the user — don't guess. Which list applies depends on whether the output is `spec.md`, `design.md`, or both — ask the user up front if it's not obvious from the request.

For `<slice>.spec.md` (logic layer):
- **Intent** — why does this slice exist, in a sentence or two of domain language (not an implementation summary)?
- **Scope & non-goals** — what's explicitly *not* this slice's job, especially anything a future contributor could plausibly assume belongs here?
- **Invariants** — what must hold no matter how the implementation changes later? This is usually the highest-value part of the spec: it's the part that actually gets checked against future changes.
- **Deferred / open questions** — anything intentionally left undecided, and what event or threshold would trigger deciding it (a scaling point, a feature-priority call, a compliance requirement)?
- **Acceptance criteria** — how would someone verify this spec is satisfied — specific tests, a manual check, or both?
- **Context (optional)** — only if Step 1's investigation turned up domain background or a mechanism explanation that doesn't fit any field above and that a future reader would need — most slices don't need this; don't ask about it otherwise.

For `<slice>.design.md` (presentation layer, only for a slice with a UI-facing surface):
- **Scope** — which screen(s)/surface(s) does this cover, and what's deliberately deferred to a later design pass?
- **Layout** — what's actually on the surface and how is it structured, beyond what one mockup snapshot shows?
- **States** — every state the surface must handle (loading, empty, error, success, and any domain-specific one) and what each shows — a mockup usually only shows the happy path.
- **Interactions** — what can the user do, what happens, and which actions need a confirmation step before firing a logic-layer consequence (check the slice's `spec.md` if one exists, don't re-decide an invariant here).
- **Consumes** — what data/endpoints this surface reads and writes, mapped back to `spec.md`'s contract where one exists.
- **Deferred / open questions** and **Acceptance criteria** — same meaning as the spec-side versions above, aimed at the presentation layer.
- **Context (optional)** — same meaning as the spec-side version above, aimed at presentation-layer theory (e.g. why a third-party component behaves the way it does) — most slices don't need this.

Ask in a short, focused batch (`AskUserQuestion` when the choices are concrete and enumerable) rather than dumping the whole list at once — a spec or design built from five rushed answers is worse than one built from two well-considered ones, continued next turn if needed.

## Step 3 — Reconcile conflicts explicitly

Any mismatch found in Step 1, or any place where the user's answer contradicts an existing doc/rule/the slice's own current spec or design, gets surfaced and resolved with the user before it goes anywhere — never silently pick the code's version or the doc's version. This includes a spec/design mismatch — e.g. an interaction described in `design.md` implying a rule `spec.md` doesn't actually state — surface it rather than letting the design silently invent logic-layer behavior. Once resolved, note which side won and why; that resolution is what gets folded into the updated file(s) — this system doesn't keep a separate decision log for it.

Any rule candidate flagged in Step 1 gets confirmed here too — ask the user whether the pattern is deliberate and worth enforcing on every future touch of matching files, not just incidental consistency in the code read so far. Only a confirmed candidate goes into Step 4's handoff; drop an unconfirmed one rather than passing it along as a maybe.

## Step 4 — Hand off to archivist

Once the relevant fields (spec: intent, scope, contract, invariants, deferred questions, acceptance criteria, optional context — or design: scope, layout, states, interactions, consumes, deferred questions, acceptance criteria, optional context) are each either confirmed from code/mockup or answered by the user, package the result and invoke `archivist` to write it: a new or updated `<slice>.spec.md` via `references/spec.template.md`, a new or updated `<slice>.design.md` via `references/design.template.md` if the slice has a UI-facing surface in scope, and any rule candidate confirmed in Step 3 for `archivist` to write separately as its own `.claude/rules/<topic>.md` — a cross-cutting convention doesn't belong inside a slice's own files. This skill's job ends at producing that material — it never writes into `deltas/` or `.claude/rules/` itself, even when the answer seems obvious enough to just write down directly.

If the resulting spec/design implies nontrivial implementation work, that's ordinary next-step coding work, not a deliverable this skill (or `archivist`) produces — sequencing it is left entirely to however the session gets implementation done.

---

## Non-goals

- Don't re-ask about anything already settled in the slice's current `spec.md` or `design.md`, if it has them — Step 0 exists specifically to prevent that.
- Don't write `spec.md` or `design.md` directly, even to save a round-trip — the separation from `archivist` is what keeps every spec's shape consistent, not a formality to skip when it's inconvenient.
- Don't force every field to a firm answer before handing off. "This is intentionally undecided until X" is a complete, valid answer for Deferred / Open questions — manufacturing a premature decision just to fill the section produces a spec/design that lies about its own certainty.
- Don't spec out a slice that's a single trivial file with no invariant worth protecting — this system's ceremony pays for itself against real complexity, not by default. Same bar applies to `design.md`: don't create one for a trivial surface with no real state/interaction complexity worth pinning down.