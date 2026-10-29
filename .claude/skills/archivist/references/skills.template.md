<!--
Template for .claude/skills/<name>/SKILL.md — a capability the main session or
an agent pulls in across tasks. Fill every section; delete guidance comments
before presenting the draft.
How a skill loads: name and description are always in context, the body loads
when it triggers, and scripts/, references/ and assets/ load on demand.
- description does double duty: what the skill does AND when to use it. Lean
  slightly pushy, because skills tend to under-trigger: spell out the phrasings
  and contexts.
- Keep the body under the skill-body ceiling in the context-budget config. Past
  it, move material not needed on every invocation to references/*.md, with an
  explicit step in the body saying when to read it.
- Write in the imperative and explain why, not bare musts.
- This skeleton gives the shape only. The content comes from how the task is
  really done in this project (Step 0): check existing docs, rules and agents so
  the skill doesn't duplicate or contradict them.
- The volatile-vs-durable rule (SKILL.md Step 2) covers the body and its
  references: point at a durable structure or describe the pattern, never cite
  an implementation file as proof.
-->
---
name: <kebab-case-name>
description: <What the skill does AND when to use it, with the concrete phrasings and contexts that should trigger it.>
---

# <Name>

<!-- One short paragraph: what makes this skill different from doing the task ad hoc, and what its output or end state is. -->

---

## Workflow

### 1. <First step>

<!-- Imperative voice, explain why not just what. -->

### 2. <Next step>

---

## Non-goals

<!-- What NOT to do even though it might seem in scope: the thing that keeps this skill from creeping into a different one. -->
