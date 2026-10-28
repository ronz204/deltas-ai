---
paths:
  - ".claude/**"
  - "CLAUDE.md"
  - "deltas/**"
---

# Context Budget Conventions

This governs how much context each knowledge-base artifact may cost, and what to do when one outgrows its ceiling. It applies to `CLAUDE.md`, docs, rules, skills, subagents, specs, and designs. The ceilings themselves live in the project's context-budget config, not here, so there is one place to change them. Whether a smaller knowledge base makes an agent more accurate is not claimed here; this rule only keeps the cost of loading it bounded and visible.

---

## Cost is size times how often it loads

- Every artifact belongs to one load class: always loaded (`CLAUDE.md`, rules without `paths:`, skill and subagent descriptions), loaded when a matching file is read (rules with `paths:`), loaded when invoked (a skill or subagent body), or loaded only on demand (docs, references, specs, designs). The same number of lines is far cheaper on demand than always loaded, because loaded context is re-read on every later turn of the session.
- Ceilings are therefore set per class, and the always-loaded class also has a total. A single always-loaded line costs more than a whole on-demand file's worth of lines over a long session.

## When an artifact is over its ceiling

- Do not ship it over, and do not trim it to fit. Trimming silently deletes information someone wrote on purpose; shipping it over lets the cost grow unnoticed.
- Fit it by splitting: what is used on every load stays, what is used only sometimes moves to something loaded on demand and is replaced by an explicit pointer that says when to read it. The mechanics per artifact type are in the archivist skill's split guide.
- Split by how often content is used, not by how big it is. Moving content that every invocation needs only adds a read step without saving anything.
- A split is lossless and confirmed: the plan (what stays, what moves where, which pointer replaces it) is shown to the user first, and every moved piece must stay reachable from a pointer that resolves.

## What is never split mechanically

- A skill or subagent description: it is always loaded and it is what triggers the skill, so shortening it risks the skill not firing. Tighten it by hand, keep the trigger phrasings that matter, and confirm with the user.
- A spec: one that is over its ceiling usually spans more than one capability, which is a scoping question for the `specifier` skill, not a file-splitting one. A slice is never split by layer.

## Measuring

- The budget script bundled with the archivist skill reads the config, classes every file by how it loads, and reports overruns and broken pointers. It only reads, so running it costs no model tokens. Token figures are estimates (about 3.3 characters per token, calibrated against Claude Code's own `/context` report); line and character counts are exact.

---

## Non-goals

- Does not decide what is worth documenting; a small artifact is not automatically a good one.
- Does not set the right ceilings from evidence. The starting values are reasoned estimates and are expected to be adjusted once real overruns show where they are too tight or too loose.
