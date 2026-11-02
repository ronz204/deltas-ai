# {{.Project}}

<!-- 2-3 lines: what this project is and what it does, from a technical point of view. -->

---

## How to work

The person never types `delta` commands. They either ask in plain words or call a skill by name. For any plain request that adds, changes, removes, checks or documents behavior, use the `comper` skill first: it reads `delta status` and hands the work to the right skill. `/comper` alone says which skill comes next.

| The state is | The next step is |
|---|---|
| No spec covers the behavior, or the contract must change | `/drafter <slice>` writes it (draft) |
| A draft | Ask the person in plain words whether to approve it; on a yes, run `delta approve <slice>` |
| Approved or stale | `/builder <slice>` builds it, tests first |
| Built | `/auditor <slice>` verifies it |
| The project needs explaining | `/scriber <doc>` writes a doc for people |

Talk to the person in their language; write everything the project keeps in English.

---

## Where to look

| If you need... | Read |
|---|---|
| The contract of a capability before changing its behavior | `deltas/specs/`: list the folder and open the spec that matches |
| Which contracts are stale or unverified | `delta status` |
| Why the project is built the way it is, or what a domain term means | `deltas/docs/`: written for people, so open one only when the task is about those concepts |
| A convention for the files you are editing | Nothing to do: `.claude/rules/` loads by itself when a matching file is read |

## Commands

<!-- Only what a session could not guess from the manifest or CI config. Omit this section if there is nothing like that. -->

## Non-goals

<!-- Optional: a scope boundary of the project itself that is easy to get wrong by accident. -->
