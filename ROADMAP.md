<!--
Template for a project's ROADMAP.md — a linear, step-by-step progress tracker.
Copy this file into a project's root and fill it in with that project's real
steps and tasks. Delete this guidance comment once filled.

Mechanics:
- Each step is one ordered phase of the project, broken into concrete tasks.
- A step's % is always checked-tasks / total-tasks for that step — never
  hand-typed. Recompute it whenever a task is checked or added.
- Overall progress at the top is checked-tasks / total-tasks across every
  step combined — same rule, no separate estimate.
- Never mark a step "Done" while it still has an unchecked task under it —
  split the step instead of rounding up.
- Keep steps linear and sequential (the order they actually happen in), not
  a dependency graph — this format is for perceived progress, not scheduling.
-->

# Roadmap

<One sentence: what this roadmap tracks — this project's build progress, not a changelog or a decision log.>

---

## Overview

**Overall progress: `<checked>` / `<total>` tasks — `<percent>`%**

| # | Step | Status | Progress |
|---|---|---|---|
| 1 | [`<Step 1 name>`](#1--step-1-name) | 🔲 Not started | 0/`<n>` — 0% |
| 2 | [`<Step 2 name>`](#2--step-2-name) | 🔲 Not started | 0/`<n>` — 0% |
| 3 | [`<Step 3 name>`](#3--step-3-name) | 🔲 Not started | 0/`<n>` — 0% |

Status legend: 🔲 Not started · 🔄 In progress · ✅ Done (only once every task under it is checked).

---

## 1 — `<Step 1 name>`

<!-- Optional one-liner: what this step actually accomplishes. -->

- [ ] `<Task>`
- [ ] `<Task>`

## 2 — `<Step 2 name>`

- [ ] `<Task>`
- [ ] `<Task>`

## 3 — `<Step 3 name>`

- [ ] `<Task>`
- [ ] `<Task>`

---

## Adding a new step

Append a new `## N — <name>` section with its own task checklist, add its row to the Overview table, and recompute **Overall** as the new total checked/total across every step. Never mark a step "Done" while an unchecked task remains under it — split the step instead of rounding up.
