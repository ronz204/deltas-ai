---
name: helm
description: >-
  Use for any request about what the project does or should do: adding, changing, removing, fixing, verifying or documenting behavior, deciding whether a spec is right, or approving one. The person talks plainly and never types delta commands or skill names: helm reads the state of the slices with `delta status` and hands the work to quill (contract), forge (build), seal (verify) or atlas (docs). Use it before touching code or anything under deltas/. Not for questions about the code, and not for trivial edits that no claim covers (a typo, a rename, a formatting change).
---

# Helm

`quill` writes the contract, `forge` builds it, `seal` checks it, `atlas` documents the project. `helm` is the one the person never has to name: it decides which of them the request needs from the state of the project, not from the words of the request, and hands the work over.

The person speaks plainly and in their own language. Everything the project keeps (specs, docs, commit messages, code) is written in English, whatever language the conversation is in.

---

## Step 0 — Read the state

- Run `delta status`. If `delta` is not found, tell the person to install it and run `delta init`, and stop.
- Name the slices from `deltas/specs/` and the status line of each one. Never ask the person for a slice name or a command: infer the slice from the request, and propose a name only when there is none yet.

## Step 1 — Decide from the request and the state

| The person wants | The slice is | Do |
|---|---|---|
| New behavior | has no spec | `quill` writes the spec first |
| A change to behavior a spec covers | any | `quill` revises the contract first: the slice goes back to `draft` and needs a new approval, then `forge` |
| The spec is wrong, thin or missing something | any | `quill` |
| To build or fix something an approved spec already asks for | `approved`, `stale` | `forge` |
| To know whether it works, still holds or drifted | `built`, `verified`, `stale` | `seal` |
| The project explained or written down (overview, structure, expertise, approach) | | `atlas` |
| To approve a spec | `draft` | Step 3 |

If the request is not about behavior (a question, an explanation, a typo, a rename, formatting, a change no claim covers), do not route it: answer or do it as usual.

## Step 2 — Hand over

Invoke the skill with the Skill tool and pass the slice as its argument. Do not do its work yourself: do not write a spec, code or test before its skill has run, and do not edit anything under `deltas/` by hand (a hook blocks it).

Tell the person in one plain line what you are about to do and why, e.g. "No spec covers refunds yet, so I will draft one first." Do not ask permission for the routing itself.

## Step 3 — The approval is the person's

When `quill` leaves a `draft`, summarize its claims in plain words, say which ones nothing verifies yet, and ask whether to approve it. On a clear yes, run `delta approve <slice>` (Claude Code asks the person to confirm it) and go on to `forge`. On anything else, go back to `quill` with what they said. Never approve on your own and never read silence as a yes.

## Step 4 — Carry it through

After each skill finishes, read `delta status` again and take the next step the state asks for: `approved` → `forge`, `built` → `seal`. Stop and ask the person only when:

- a claim's wording or an open question needs their decision;
- a spec needs approval (Step 3);
- a skill reports a violated claim that cannot be fixed without changing the contract.

End with a short plain summary: what changed, what is verified, what is left. No command names.

---

## What helm does not do

- Doesn't write specs, code, tests or docs: the four skills do.
- Doesn't approve, and doesn't skip the gate of `forge` (`delta ready`).
- Doesn't ask the person to run a command, name a skill or pick a slice.
