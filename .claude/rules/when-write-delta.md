---
paths:
  - "*/deltas/**"
  - ".claude/skills/specifier/**"
  - ".claude/skills/archivist/**"
---

# Delta Scope Conventions

This governs how to decide whether something deserves its own `<slice>.spec.md`, and how to size that slice, once the underlying work spans multiple architectural layers (domain, application, infrastructure) or multiple bounded contexts — the recurring source of "I don't know when this needs a delta."

---

## Scope a spec to a capability, never to a layer

- A slice is sized the way a bounded context (or a cross-context workflow, see below) would be sized — never the way an architecture layer would be sized. Don't name or split a spec by layer (e.g. one spec for a bounded context's domain logic, a second for its application-layer orchestration, a third for its infrastructure), because a layer isn't a capability boundary — it's an implementation facet of realizing one, and splitting along it produces specs whose invariants keep changing together for the same reason, which is the same problem as two slices that should really be one.
- The litmus test for "does this deserve its own spec": does it have invariants or a contract that could later be checked against the implementation, and that are genuinely distinct from any existing slice's? A capability that only restates another slice's contract from a different angle doesn't clear this bar.

## When it isn't a spec at all

- A recurring technical pattern applied across multiple bounded contexts — how repository adapters are implemented, an eventing/outbox convention, an auth middleware convention — is a rule candidate, not a spec: it's "how we build things," not a capability's own contract.
- A static infrastructure fact — hosting model, database engine, provisioned resources — belongs in the project's own structure/persistence reference, not a spec: there's no capability being contracted, only a fact to record.
- A slice-specific consequence of infrastructure or security (e.g. "must tolerate at-least-once delivery from a given queue without duplicating effects") is an Invariant inside that slice's own `spec.md`, not a reason to create a separate infrastructure- or security-flavored spec.

## Application-layer orchestration

- A use case or command that only orchestrates a single bounded context doesn't need its own spec — that orchestration is part of the owning bounded context's contract; name the use case/command as part of its Contract section instead of splitting it out.
- A workflow that orchestrates more than one bounded context earns its own spec only when the orchestration itself has invariants none of the underlying bounded contexts have alone — ordering guarantees, compensation/rollback on partial failure, idempotency across the boundary. When it does, name that spec after the workflow itself, never after a layer, and have it reference the underlying bounded contexts' `spec.md` contracts rather than restating their invariants — the same discipline `design.md` already applies toward `spec.md`, so the two don't drift against each other.

---

## Non-goals

- Doesn't decide where one bounded context's boundary ends and another's begins, or which domain concepts belong to which context — that's a per-bounded-context decision made in its own `spec.md`.
- Doesn't apply to a project without multiple architectural layers or bounded contexts — the ambiguity this resolves only exists once there's more than one layer or context to potentially split across.
