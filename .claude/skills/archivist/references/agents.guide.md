# Writing subagents

Loaded by archivist when writing `.claude/agents/<name>.md`; skeleton in `agents.template.md`. Step numbers below refer to archivist's own SKILL.md.

A subagent is a markdown file with YAML frontmatter; the body is its system prompt. It runs in its own context window with its own tool access — use it for a bounded, repeatable task you'd otherwise delegate the same way every time.

```yaml
---
name: spec-invariant-checker
description: Reviews changes to a slice's implementation against the invariants declared in its <slice>.spec.md before merge. Use after editing source files inside a slice that has a spec.
tools: Read, Glob, Grep
model: sonnet
---

You are checking a diff against the invariants declared in <slice>.spec.md for the slice it touches. For each invariant listed:
- confirm the changed code still upholds it
- if an invariant is no longer upheld, say so explicitly — don't decide on the user's behalf whether that's acceptable
- ...
```

- **`description` is the trigger** — same rule as skills: be specific about when to invoke it, don't undersell it.
- **Restrict `tools` to the minimum** the task needs — a review/analysis agent shouldn't get write access it doesn't use.
- **Choose `model` deliberately**: cheaper/faster models for narrow search or lint-style checks, the default for anything requiring judgment.
- **Ground the system prompt in this project's actual conventions** rather than generic reviewer advice — pull specifics from `.claude/rules/` and `.claude/docs/` rather than restating boilerplate. Same volatile-vs-durable reference rule as Step 2: describe the pattern the agent should check for, never cite a specific file as the example of it.
- Names must be unique across `.claude/agents/` — a collision gets silently discarded, not merged.
- Use `references/agents.template.md` as the fill-in skeleton; the example above shows the shape it produces.
