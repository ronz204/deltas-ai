---
paths:
  - ".claude/CLAUDE.md"
  - ".claude/docs/**"
  - ".claude/rules/**"
  - ".claude/agents/**"
  - ".claude/skills/**"
  - "*/deltas/**"
---

# Delta Knowledge-Base Editing Conventions

These paths are the durable knowledge base this project's delta methodology maintains — `CLAUDE.md`, `.claude/docs/`, `.claude/rules/`, `.claude/agents/`, `.claude/skills/`, and every slice's `spec.md`/`design.md` under `deltas/` at the project root (or, in a repo genuinely split into multiple services, under each service's own `deltas/`). They exist to stay grounded and current; editing them ad hoc is exactly the kind of drift the methodology is meant to prevent.

---

## Route through the matching skill, don't hand-edit

- Writing or updating `CLAUDE.md`, a doc, a rule, an agent, a skill, or a slice's `spec.md`/`design.md` goes through `archivist`, because it's the one place that knows which shape/template each of these takes and enforces the volatile-reference and writing conventions consistently. A hand-edit skips both.
- A new or substantially-revised `<slice>.spec.md`/`<slice>.design.md` goes through `specifier` first, because intent, invariants, and states/interactions have to come from investigation and user Q&A, not be invented while editing the file. `archivist` still does the actual write once that material exists.
- Checking whether a slice's `spec.md`/`design.md` still matches the real implementation goes through `sentinel`, which reports drift — it never edits the file itself; a correction it surfaces still goes back through `archivist`.

## Trivial edits are exempt

A one-line fix to something already fully settled — a typo, a stale command that changed — doesn't need the full pipeline. The bar is whether the content is already known and uncontested, not the size of the diff; anything that requires deciding, inferring, or interviewing still routes through the skill that owns that judgment.

---

## Non-goals

- Doesn't apply outside these paths — editing application source, tests, or config the knowledge base merely describes is unaffected.
