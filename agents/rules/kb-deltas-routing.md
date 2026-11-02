---
paths:
  - "CLAUDE.md"
  - ".claude/CLAUDE.md"
  - ".claude/docs/**"
  - ".claude/rules/**"
  - ".claude/agents/**"
  - ".claude/skills/**"
  - "deltas/**"
  - "**/deltas/**"
---

# Delta Knowledge-Base Editing Conventions

These paths are the durable knowledge base this project's delta methodology maintains — `CLAUDE.md`, `.claude/docs/`, `.claude/rules/`, `.claude/agents/`, `.claude/skills/`, and every slice's `<slice>.spec.md` under `deltas/specs/` (the project has a single `deltas/`, at its root, even when it is split into services). They exist to stay grounded and current; editing them ad hoc is exactly the kind of drift the methodology is meant to prevent.

---

## Route through the matching skill, don't hand-edit

- A hook blocks the agent's `Write` and `Edit` on these paths. A person can edit `CLAUDE.md`, docs, rules, agents and skills by hand; the agent proposes the change and the person makes it.
- A new or substantially-revised `<slice>.spec.md` goes through `drafter`, because intent and invariants have to come from investigation and user Q&A, not be invented while editing the file. `drafter` writes it with `delta spec write`, which validates before it saves; the result is a `draft` that a person approves with `delta approve`.
- Linking a test to a claim is `delta link`, which only changes that claim's tag. Moving a slice along its states is `delta approve`, `delta built` and `delta check`.
- Checking whether a slice still matches the real implementation goes through `auditor`, which reports drift — it never edits the file itself; a correction it surfaces goes back through `drafter`.

## Trivial edits are exempt

A one-line fix to something already fully settled — a typo, a stale command that changed — doesn't need the full pipeline. The bar is whether the content is already known and uncontested, not the size of the diff; anything that requires deciding, inferring, or interviewing still routes through the skill that owns that judgment.

---

## Non-goals

- Doesn't apply outside these paths — editing application source, tests, or config the knowledge base merely describes is unaffected.
