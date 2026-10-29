<!--
Template for CLAUDE.md — the project's root orientation file, loaded into every
session automatically. Locate the existing file before assuming a path (repo
root or .claude/) so a second one isn't created elsewhere.
It is a router, not a reference: its job is to get a session oriented and point
to the right deeper file, never to hold a second copy of what it points to. It
is the one artifact allowed (expected) to reference other files by path.
Keep it short: every line costs context on every turn. Claude Code itself warns
past 200 lines; this project's ceiling is in the context-budget config.
- Write the index as questions ("if you need X, read Y"), not an inventory.
- Leave out what the code or tool config already shows (directory listings,
  dependency lists, obvious commands): that is the content that drifts.
- Don't use `@path` imports to organize it: imported files load together with
  this one, so they add context instead of saving it.
Fill every section; delete guidance comments before presenting the draft. Ground
every claim in the real package manifest, build config, CI files and tree.
-->

# <Project name>

<!-- 2-3 lines: what this project is and what it does. Technical POV — no pitch, no ROI framing, no competitive positioning, even if that's how the team talks about it elsewhere. -->

---

## Where to look

<!--
The index. One row per question a session actually has; the second column is a
trigger ("read X when you need Y"), not a description of the folder. List only
what actually exists in this project — no row for a directory that doesn't
exist yet, none for a convention this project hasn't adopted — and keep a
doc row only for a doc that exists (a category with nothing to say isn't created). The `deltas/` row is the default
for a project with one root; if this repo is genuinely split into multiple
independently-deployable services, replace it with one row per service's own
`<service>/deltas/`. Don't list the slices themselves: the session finds them
by listing the folder, so adding one never means editing this file.
-->

| If you need... | Read |
|---|---|
| The contract of a capability before changing its behavior | `deltas/` — list the folder and open the spec that matches (and its design file for UI work) |
| The project's scope, non-goals and domain vocabulary | `.claude/docs/overview.md` |
| Behavior that spans components or isn't evident from one | `.claude/docs/modules.md` |
| Why the architecture is the way it is: decisions, infrastructure, cross-cutting patterns | `.claude/docs/structure.md` |
| The data model and persistence | `.claude/docs/database.md` |
| How a non-trivial mechanism works | `.claude/docs/expertise.md` |
| A repeatable task or capability | `.claude/agents/` and `.claude/skills/` |
| What the agent may or may not run | `.claude/settings.json` — <one sentence on the reasoning only if the file doesn't make it obvious> |
| A convention for the files you're editing | Nothing to do: `.claude/rules/` loads by itself when a matching file is read |

## Commands

<!--
Only what a session could not guess from the package manifest, task runner or
CI config: a required environment variable, a non-default test invocation, a
step order that matters. Confirm each against the real file; never write a
command that "should" work. Omit the whole section if there is nothing like
that — a command the manifest already shows doesn't belong here.
-->

| Task | Command |
|---|---|
| <what it does> | `<command>` |

## Non-obvious structure

<!--
Only what the file tree doesn't show: generated or vendored directories not to
edit, a boundary between deployable units, an entry point that isn't where you'd
look. Read the real tree before writing. Omit the whole section if the tree
speaks for itself.
-->

## Conventions

<!-- Point at .claude/rules/ for anything conditionally-loaded; state here only what applies globally and isn't already its own rule file — e.g. a project-wide non-negotiable. Keep this short — a growing list belongs in a new rule, not here. -->

---

## Non-goals

<!-- Optional. Include only if there's a scope boundary for the project itself (not the knowledge base) that's easy to get wrong by accident. -->
