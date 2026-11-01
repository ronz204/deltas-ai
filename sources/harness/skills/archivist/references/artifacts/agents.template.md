<!--
Template for .claude/agents/<name>.md — a subagent: a bounded, repeatable task
with its own system prompt, its own tool access and its own context window. Use
one for work you would otherwise delegate the same way every time. Fill every
section; delete guidance comments before presenting the draft.
- description is the trigger, same discipline as a skill's: say what it does AND
  when to invoke it, with a concrete "use after/before X" cue. Don't undersell it.
- tools: the minimum the task needs. A review agent gets no write access it
  won't use.
- model: choose deliberately. Cheaper and faster for narrow search or lint-style
  checks, the default for anything that needs judgment.
- Names are unique across .claude/agents/; a collision is silently discarded,
  not merged.
- Ground the system prompt in this project's conventions: pull specifics from
  .claude/rules/ and .claude/docs/, not generic reviewer advice. Describe the
  pattern to check for, never cite a file as its example (SKILL.md Step 2).
-->
---
name: <kebab-case-name>
description: <What this agent does AND when to invoke it, with a concrete "use after/before X" cue rather than a vague summary.>
tools: <Minimum tool list this task needs, e.g. "Read, Glob, Grep" for a review-only agent.>
model: <Chosen deliberately, see above.>
---

You are <role, one sentence — what this agent is responsible for and the boundary of that responsibility>. Check specifically for:
- <concrete, checkable criterion 1>
- <concrete, checkable criterion 2>
- <...>

<!-- State what the agent does when it finds a violation (flag it, fix it inline, refuse, ask); don't leave the output behavior implicit. -->
