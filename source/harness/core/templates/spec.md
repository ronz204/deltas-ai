---
format: 1
slice: {{.Slice}}
status: draft
---

# {{.Slice}}: spec

> One line: what this slice is responsible for.

## Intent

<!-- Why this slice exists, in the domain's own vocabulary, not an implementation summary. -->

## Scope

<!-- What this slice owns, then its non-goals. A boundary that is easy to cross by accident earns a line. -->

## Contract

<!-- The shape a caller can rely on: inputs, outputs, behavior guarantees. A table or code block beats prose for structure. -->

## Invariants

<!-- One claim per invariant: a header "- INV-01 [tag]" and, indented under it, one EARS sentence ("When <trigger>, <subject> <response>"; also While, Where, If...then). Tags: test: <command>, manual, none. -->

## Deferred / Open questions

<!-- Decisions not made yet and what would trigger them. A claim tagged none is a testing gap: list it here. -->

## Acceptance criteria

<!-- One claim per criterion: a header "- ACC-01 [tag]" and, indented under it, one "Given <context>, when <action>, then <outcome>" sentence. List only what the invariants do not already cover. -->
