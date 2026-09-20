<!--
Template for .cursor/rules/orientation.mdc — the project's root orientation
rule, loaded into every Cursor session automatically via alwaysApply: true.
This is the one artifact in this system that's allowed — expected — to
reference other files by path; its entire job is routing to where the real
content lives, not holding that content itself. Keep it short: every line
here costs context on every single turn, regardless of whether that turn
needs it — aim for under ~200 words, since this rule shares Cursor's
recommended ~2,000-token combined budget for all alwaysApply rules in the
project.
Fill every section; delete guidance comments before presenting the draft.
Ground every claim the same as any other artifact — read the actual package
manifest / build config / CI files for commands and layout, don't guess.
-->
---
alwaysApply: true
---

# <Project name>

<!-- 2-3 lines: what this project is and what it does. Technical POV — no pitch, no ROI framing, no competitive positioning, even if that's how the team talks about it elsewhere. -->

---

## Knowledge base layout

<!--
Table explaining what lives under .cursor/ (and wherever specs live) and
why — this is the map a session uses to find durable context without
re-deriving it from the code. List only what actually exists in this
project; don't include a row for a directory that doesn't exist yet, and
don't invent a row for a convention this project hasn't adopted. The
`deltas/` row below is the default for a project with one root — if this
repo is genuinely split into multiple independently-deployable services,
replace it with one row per service's own `<service>/deltas/` instead.
-->

| Path | Holds |
|---|---|
| `.cursor/docs/` | <self-contained reference files, one concern each — e.g. architecture, module reference> |
| `.cursor/rules/` | <conventions loaded automatically, either always, on a matching file (`globs`), or when the agent judges them relevant (`description`)> |
| `.cursor/skills/` | <capabilities pulled in across tasks, e.g. this project's own archivist/specifier pair> |
| `deltas/` | <per-slice spec/design files: `<slice>.spec.md`, optional `<slice>.design.md`> |

## Repo layout

<!--
How the actual codebase is composed — top-level directories and what each
is for. This is about the code, not the knowledge base above. Read the
real tree before writing this; don't infer it from the project name or
from what a similar project usually looks like.
-->

| Path | Purpose |
|---|---|
| `<dir>` | <what lives here> |

## Setup & common commands

<!--
Pull these from the actual package manifest / task runner / CI config —
never write a command that "should" work without confirming it does. Omit
a row for anything this project doesn't have (e.g. no separate lint step).
-->

| Task | Command |
|---|---|
| Install dependencies | `<command>` |
| Run in development | `<command>` |
| Run tests | `<command>` |
| Lint / typecheck | `<command>` |
| Build | `<command>` |

## Conventions

<!-- Point at .cursor/rules/ for anything conditionally-loaded; state here only what applies globally and isn't already its own rule file — e.g. a project-wide non-negotiable. Keep this short — a growing list belongs in a new rule, not here. -->

---

## Non-goals

<!-- Optional. Include only if there's a scope boundary for the project itself (not the knowledge base) that's easy to get wrong by accident. -->
