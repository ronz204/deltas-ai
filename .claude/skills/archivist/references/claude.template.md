<!--
Template for CLAUDE.md — the project's root orientation file, loaded into
every Claude Code session automatically. This is the one artifact in this
system that's allowed — expected — to reference other files by path; its
entire job is routing to where the real content lives, not holding that
content itself. Keep it short: every line here costs context on every
single turn, and Claude Code itself warns past 200 lines.
Write it as an index of questions ("if you need X, read Y"), not as an
inventory of what exists. Leave out anything the agent can derive from the
code or the tool config — directory listings, dependency lists, commands a
package manifest already makes obvious — because that is exactly the content
that drifts and that distracts. Don't use `@path` imports to organize it:
imported files load together with this one, so they add context instead of
saving it.
Fill every section; delete guidance comments before presenting the draft.
Ground every claim the same as any other artifact — read the actual package
manifest / build config / CI files and the real tree, don't guess.
-->

# <Project name>

<!-- 2-3 lines: what this project is and what it does. Technical POV — no pitch, no ROI framing, no competitive positioning, even if that's how the team talks about it elsewhere. -->

---

## Where to look

<!--
The index. One row per question a session actually has; the second column is a
trigger ("read X when you need Y"), not a description of the folder. List only
what actually exists in this project — no row for a directory that doesn't
exist yet, none for a convention this project hasn't adopted — and adapt the
doc rows to the docs the project really has. The `deltas/` row is the default
for a project with one root; if this repo is genuinely split into multiple
independently-deployable services, replace it with one row per service's own
`<service>/deltas/`. Don't list the slices themselves: the session finds them
by listing the folder, so adding one never means editing this file.
-->

| If you need... | Read |
|---|---|
| The contract of a capability before changing its behavior | `deltas/` — list the folder and open the spec that matches (and its design file for UI work) |
| What the system is, its scope and vocabulary | `.claude/docs/overview.md` |
| Behavior that spans components or isn't evident from one | `.claude/docs/modules.md` |
| The stack, topology, how the parts communicate | `.claude/docs/structure.md` |
| The data model and persistence | `.claude/docs/database.md` |
| How a non-trivial mechanism works | `.claude/docs/expertise.md` |
| The build order, roadmap and done criteria | `.claude/docs/approach.md` |
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
