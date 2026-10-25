# Writing new skills

Loaded by archivist when writing `.claude/skills/<name>/SKILL.md`; skeleton in `skills.template.md`. Step numbers below refer to archivist's own SKILL.md.

A skill's `name` + `description` are always in context; the body loads only when it triggers; anything under `scripts/`, `references/`, or `assets/` loads on demand.

- **`description` does double duty**: state what the skill does *and* when to use it, and lean slightly pushy — Claude tends to under-trigger skills, so spell out phrasings and contexts explicitly rather than trusting a vague description to be inferred.
- Keep the `SKILL.md` body under the skill-body ceiling in the context-budget config. If it's growing past that, split material that isn't needed on every invocation into `references/*.md` and point to it from the body with an explicit step that says when to read it, rather than inlining everything.
- Write instructions in the imperative, and explain *why* rather than issuing bare musts — same principle as docs and rules.
- Every new skill goes through the same Step 0 discipline: don't invent its instructions from a generic template — `references/skills.template.md` gives the *shape*, but the actual content still has to come from how the task is really done in this project, checking existing docs/rules/agents first so the new skill doesn't duplicate or contradict what already exists.
- Same volatile-vs-durable reference rule as Step 2 applies to `SKILL.md` and everything under its `references/` — a skill's instructions and reference material should point at a durable structure (a project, a directory holding one concern) or describe a pattern in the abstract, never cite a specific implementation file as the proof that pattern exists in this repo.
