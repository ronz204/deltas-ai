<!--
Template for deltas/<slice>.design.md (at the project root) — the living,
current UI-facing contract for a slice: the presentation-layer counterpart
to <slice>.spec.md's logic-layer contract. Created only for slices that have
a UI-facing surface — most slices never need this file.
Edited in place as the slice's presentation contract changes; this file is
always current truth, not a log of how it got there. Never restate an
invariant or business rule that already lives in spec.md — reference the
part of its contract that's relevant to rendering this surface instead of
duplicating it, or the two files will drift against each other.
In a repo genuinely split across multiple services, this file lives under
whichever service actually renders the UI, in that service's own
deltas/ — which may not be the same service <slice>.spec.md lives under. If
so, name that other service explicitly wherever this file points at the
logic contract, since a reader won't have both files in the same directory.
Fill every section; delete guidance comments before presenting the draft.
Ground every claim in the actual source implementing the slice's UI once it
exists, and in the project's design-system reference if one exists — never
write this from a mockup's own internal prototyping wiring, and never from
memory.
-->

# <Slice name> — Design

> <One line: what surface(s) this covers, and which slice's spec.md it's the presentation layer for.>

## Scope

<!-- What this design owns — which screen(s)/surface(s), which states. Then explicit non-goals — a surface or breakpoint intentionally left for a later pass earns a line here. -->

## Layout

<!-- The concrete shape of the surface: structure, composition, key regions. Prefer a short structural breakdown over prose — this describes arrangement, not marketing copy. Point at the project's design-system reference for tokens/components instead of re-deriving them here. -->

## States

<!-- Every state the surface must handle explicitly: loading, empty, error, success, and any domain-specific state (partially filled, pending confirmation, etc.). What each one shows, not just that it exists. Write each as one line, e.g. "While loading, the list shows a skeleton." -->

## Interactions

<!-- What the user can do here and what happens: actions, navigation, confirmations required before a consequential action fires. Where an interaction has a logic-layer consequence, point at the relevant part of spec.md's Contract/Invariants rather than restating the rule here. Write each as one line, e.g. "When the user clicks Delete, the app asks for confirmation before deleting." -->

## Consumes

<!-- What data/endpoints/state this surface reads and writes — the presentation layer's dependency on spec.md's contract. A table of field/endpoint -> source is usually the clearest shape here. -->

## Deferred / Open questions

<!-- Presentation decisions intentionally not made yet, and what would trigger making them. An honest open question is a valid entry — don't force a premature decision just to fill this section. -->

## Acceptance criteria

<!-- How to verify this design contract is satisfied: a manual check against the real UI, a visual regression test, or both. If it can't be falsified by looking at the resulting screen, it isn't specific enough yet. -->

## Context (optional)

<!--
Omit this section entirely for the common case. Include it only when the
surface needs theory that doesn't fit any section above — e.g. why a
third-party component behaves the way it does — free-form, not a fixed
skeleton. Never a decision log; fold decision-shaped content into the
sections above as current state instead, or leave it out.
-->

---

Last updated: <date>.