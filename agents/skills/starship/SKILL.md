---
name: starship
description: >-
  Use for any request about what the project does or should do: adding, changing, removing, fixing, verifying or documenting behavior, deciding whether a spec is right, or approving one. It reads the state of the slices with `delta status` and hands the work to drafter (contract), builder (build), auditor (verify) or scribe (docs). It works two ways: the person asks in plain words and starship routes the request, or calls /starship to be told which skill comes next. Use it before touching code or anything under deltas/. Not for questions about the code, and not for trivial edits that no claim covers (a typo, a rename, a formatting change).
argument-hint: "[slice]"
---

# Starship

`drafter` writes the contract, `builder` builds it, `auditor` checks it, `scribe` documents the project. `starship` knows where the project is in that flow: it decides which of them is next from the state of the project, not from the words of the request.

It has two entrances:

- **A plain request** ("add refunds", "is billing still right?"): it routes the request, hands the work over and carries it through.
- **`/starship [slice]`** with no request: it reports the state of the slice (or of all of them) and names the skill that comes next, so the person can call it directly. It does not run that skill unless the person says so.

The person may speak plainly or call skills by name, in their own language. Everything the project keeps (specs, docs, commit messages, code) is written in English, whatever language the conversation is in.

---

## Step 0 — Read the state

- Run `delta status`. If `delta` is not found, tell the person to install it and run `delta init`, and stop.
- Name the slices from `deltas/specs/` and the status line of each one. Don't ask the person for a slice name: infer it from the request (or from `$ARGUMENTS`), and propose a name only when there is none yet.

## Step 1 — Decide from the request and the state

| The person wants | The slice is | Do |
|---|---|---|
| New behavior | has no spec | `drafter` writes the spec first |
| A change to behavior a spec covers | any | `drafter` revises the contract first: the slice goes back to `draft` and needs a new approval, then `builder` |
| The spec is wrong, thin or missing something | any | `drafter` |
| To build or fix something an approved spec already asks for | `approved`, `stale` | `builder` |
| To know whether it works, still holds or drifted | `built`, `verified`, `stale` | `auditor` |
| The project explained or written down (overview, structure, expertise, approach) | | `scribe` |
| To approve a spec | `draft` | Step 3 |

If the request is not about behavior (a question, an explanation, a typo, a rename, formatting, a change no claim covers), do not route it: answer or do it as usual.

## Step 2 — Hand over

If the person called `/starship` with no request, stop here: say in one or two lines where the slice is and which skill is next, with the call to make, e.g. "billing is approved and not built: next is `/builder billing`." Run it only if they say so.

Otherwise invoke the skill with the Skill tool and pass the slice as its argument. Do not do its work yourself: do not write a spec, code or test before its skill has run, and do not edit anything under `deltas/` by hand (a hook blocks it).

Tell the person in one plain line which skill you are calling and why, e.g. "No spec covers refunds yet, so `drafter` writes one first." Do not ask permission for the routing itself.

## Step 3 — The approval is the person's

When `drafter` leaves a `draft`, summarize its claims in plain words, say which ones nothing verifies yet, and ask whether to approve it. On a clear yes, run `delta approve <slice>` (Claude Code asks the person to confirm it) and go on to `builder`. On anything else, go back to `drafter` with what they said. Never approve on your own and never read silence as a yes.

## Step 4 — Carry it through

After each skill finishes, read `delta status` again and take the next step the state asks for: `approved` → `builder`, `built` → `auditor`. Stop and ask the person only when:

- a claim's wording or an open question needs their decision;
- a spec needs approval (Step 3);
- a skill reports a violated claim that cannot be fixed without changing the contract.

End with a short plain summary: which skills ran, what changed, what is verified, what is left, and the skill that comes next if there is one.

---

## What starship does not do

- Doesn't write specs, code, tests or docs: the four skills do.
- Doesn't approve, and doesn't skip the gate of `builder` (`delta ready`).
- Doesn't ask the person to run a `delta` command, and doesn't make them pick a slice it can infer.
