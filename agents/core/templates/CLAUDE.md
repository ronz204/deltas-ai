# {{.Project}}

<!-- 2-3 lines: what this project is and what it does, from a technical point of view. -->

---

## How to work

The person talks plainly and never types delta commands or skill names. For any request that adds, changes, removes, checks or documents behavior, use the `helm` skill first: it reads `delta status` and hands the work to the right skill.

| The state is | The next step is |
|---|---|
| No spec covers the behavior, or the contract must change | `quill` writes it (draft) |
| A draft | Ask the person in plain words whether to approve it; on a yes, run `delta approve <slice>` |
| Approved or stale | `forge` builds it, tests first |
| Built | `seal` verifies it |

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
