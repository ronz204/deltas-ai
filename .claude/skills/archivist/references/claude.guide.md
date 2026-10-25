# Writing CLAUDE.md

Loaded by archivist when writing the root orientation file; skeleton in `claude.template.md`. Step numbers below refer to archivist's own SKILL.md.

`CLAUDE.md` is the one artifact this system loads into *every* session automatically, so it behaves differently from everything else it routes to:

- **It's a router, not a reference.** Its job is to help a session get oriented fast and find the right deeper file — not to hold the actual content of what it points to. Anything that would make it a second copy of a doc belongs in the doc instead, with a pointer left behind.
- **Cross-references are expected here**, unlike `.claude/docs/*.md` — see Step 2.
- **Keep it short.** Every line costs context on every single turn, regardless of whether that turn needs it — this is the strongest length pressure of any artifact in this system. Claude Code itself recommends staying under 200 lines and warns past that.
- **An index of questions, not an inventory.** Write each row as "if you need X, read Y" so the second column is a trigger, and don't list the slices — the session finds them by listing `deltas/`, so adding one never means editing this file.
- **Leave out what the code already says.** Directory listings, dependency lists, and commands a manifest makes obvious can be derived by the session, and they are the part that drifts. Keep only what the code doesn't show: non-obvious commands, non-obvious structure, project-wide non-negotiables.
- **Don't use `@path` imports to organize it.** Imported files load together with `CLAUDE.md`, so they add context instead of saving it.
- **Locate the existing file before assuming a path.** Different projects keep it in different places (repo root, or under `.claude/`) — check before creating a second one elsewhere.
- Use `references/claude.template.md`: short project description, a "where to look" index, only the commands and structure the code doesn't make obvious, and any project-wide convention that isn't already its own rule file.
- Same Step 0 discipline applies to every factual claim in it — confirm every command against the actual package manifest/task runner/CI config and every structural claim against the actual tree, never from what a similar project usually looks like.
