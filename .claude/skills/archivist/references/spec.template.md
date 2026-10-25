<!--
Template for deltas/<slice>.spec.md (at the project root — or, in a repo
genuinely split into multiple services, under that service's own
deltas/<slice>.spec.md) — the living, current contract for one project
slice/capability's LOGIC layer (data, rules, invariants). A slice with a
UI-facing surface has a presentation-layer counterpart, <slice>.design.md
(own template, references/design.template.md) — don't fold
layout/states/interactions in here, and don't duplicate an invariant stated
here inside that file either; it should reference this one.
Edited in place as the slice evolves — this file is always the current
truth, not a log of how it got there. No append-only history is kept
alongside it; if a decision needs theory/context beyond what the contract
shape below holds, that goes in the optional Context section at the end
(free-form, no fixed skeleton), never in a decision-log entry.
Fill every section; delete guidance comments before presenting the draft.
Ground every claim in the actual source implementing the slice — never write
this from memory, from a description alone, or from an older doc's claim.
-->

# <Slice name> — Spec

> <One line: what this slice is responsible for.>

## Intent

<!-- Why this slice exists, in the domain's own vocabulary — not an implementation summary. -->

## Scope

<!-- What this slice owns. Then explicit non-goals — a boundary that's easy to violate by accident earns a line here. -->

## Contract

<!--
The concrete shape a caller/reader can rely on: what's configured/enabled and
why, inputs/outputs, behavior guarantees. Prefer a table or fenced code block
over prose wherever the shape is structural (config keys, request/response
fields, a permission matrix).
-->

## Invariants

<!--
What must always hold regardless of implementation changes — the things a future change to this slice must never break. This is usually the highest-value section: it's what actually gets checked later.
Every invariant ends with a verification tag in square brackets saying how it is checked: `[test: <path> "<test name>"]` for an automated test that exists, `[check: <read-only command>]` for a command whose result proves it (for example a grep that must find nothing), `[manual: <step>]` for a hand check, or `[none]` when nothing verifies it yet. Confirm a cited test's file and name exist before writing the tag.
Write each invariant as ONE claim in one of the five EARS forms, in plain language (no "shall"):
- always/never: "<subject> <response>", e.g. "Stock never imports from checkout."
- event: "When <trigger>, <subject> <response>", e.g. "When a percentage discount is applied, the amount rounds half up to the nearest cent."
- state: "While <state>, <subject> <response>", e.g. "While a reservation is older than 15 minutes, it does not count as reserved."
- unwanted case: "If <condition>, then <subject> <response>", e.g. "If payment fails, the stock reservation is released."
- optional feature: "Where <feature is present>, <subject> <response>", e.g. "Where a tax rate is set, tax is computed on the discounted subtotal."
The response must be something a test or a grep can observe (a returned value, an error, an emitted event, a state change, the absence of an import). If a line needs "and" to join two outcomes, split it into two invariants.
-->

## Deferred / Open questions

<!-- Decisions intentionally not made yet, and what would trigger making them. An honest open question is a valid entry — don't force a premature decision just to fill this section. An invariant tagged `[none]` is a testing gap: list it here, with what would trigger writing its test. -->

## Acceptance criteria

<!-- How to verify this spec is satisfied: specific tests, or a concrete manual check, or both. If this can't be falsified by reading the resulting code/behavior, it isn't specific enough yet. Each criterion carries the same verification tag as an invariant. List here only what the invariants don't already cover (for example the whole suite passing, or an end-to-end flow); don't restate an invariant. Write each as one Given/When/Then line, e.g. "Given a cart with two items, when checkout runs with a failing card, then no order is saved." -->

## Context (optional)

<!--
Omit this section entirely for the common case. Include it only when the
slice needs domain background or mechanism explanation that doesn't fit any
section above — free-form (a table here, a diagram there, a couple of
paragraphs), not a fixed skeleton like the rest of this file. Never a
decision log ("we changed X because Y", superseded-decision narratives,
timestamps) — fold decision-shaped content into the sections above as
current state instead, or leave it out.
-->

---

Last updated: <date>.