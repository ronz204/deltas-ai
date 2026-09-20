# Delta Methodology — Culture

How this project's knowledge base actually gets worked with day to day — the mental model behind the pipeline, not the field-by-field mechanics of any one artifact, which live with the capability that owns it.

---

## What a delta is

A delta is a bounded capability or change inside a project (or, in a repo genuinely split into multiple services, inside one of them), sized the way a bounded context would be picked — not a source file, not a ticket. Its unit of record is a contract describing what is true *now*: intent, scope, the concrete shape a caller can rely on, and the invariants that must never break. It is deliberately not a history of how that contract came to be. Time-bound narrative — "we changed X because Y," superseded-decision writeups, timestamps — is excluded on purpose: it accumulates reading cost for every future session without making the current contract any more correct. If a decision matters going forward, it is folded into the contract as current state; if it doesn't, it was never worth keeping.

This also governs how many deltas a project ends up with. Two capabilities whose contracts keep changing together for the same underlying reason are one bounded context wearing two names — merging them costs less upkeep long-term than maintaining the pair in lockstep.

## The four-role pipeline

Four capabilities cover the knowledge base's lifecycle, split by a single-responsibility line: each one's reason to change is different, so keeping them separate keeps each one's judgment consistent regardless of which capability answers on a given day.

| Stage | Responsibility | Reason to change | Writes to the knowledge base? |
|---|---|---|---|
| Bootstrap | Orient a repo with no knowledge base yet — scan or interview, breadth-first | "Did we get the project oriented fast and accurately" | No — hands material to the write stage |
| Interview | Gather what code or a mockup alone can't answer for one delta — intent, invariants, open questions, states, interactions | "Did we ask the right questions and find the right conflicts" | No — hands material to the write stage |
| Write | Land material in the correct artifact shape — the one place that knows which template each kind of file takes | "Does the artifact land in the right shape" | Yes — this is the only stage that writes |
| Verify | Check whether an already-written contract still holds against the real implementation | "Does the implementation still match its contract" | No — reports findings, hands corrections back to the write stage |

The write stage is the sole path into every artifact under this knowledge base — a doc, a rule, a subagent, a skill, the root orientation file, or a delta's own contract files. Bootstrap and interview investigate and ask; verify checks; only the write stage commits the result. A hand-edit that skips this pipeline is exactly the kind of drift the methodology exists to prevent, because nothing else re-derives the right shape or grounds the claim in the real source afterward.

The exemption is trivial edits: a one-line fix to something already fully settled — a typo, a stale command that changed — doesn't need the full pipeline. The bar is whether the content is already known and uncontested, not the size of the diff. Anything that requires deciding, inferring, or interviewing still routes through the stage that owns that judgment.

## Living contract, not a roadmap

A delta's contract file(s) — covering its logic layer, and its presentation layer where it has a UI-facing surface — are living truth: edited in place, always describing what holds right now, never allowed to go stale. This system keeps no separate roadmap file for a delta: moving from a finished contract to code is ordinary implementation work, sequenced however the session gets it done, never a second knowledge-base artifact that would need its own upkeep.

The same current-truth discipline separates a delta's own contract from the project-wide reference docs sitting above it. A project-wide doc is the cross-cutting, curated reference; a delta's contract is finer-grained, lives with the delta, and changes exactly as often as that delta does. A decision that stabilizes and starts mattering beyond one delta graduates into the project-wide reference — it does not stay duplicated in both places once that happens.

## Why there is no decision log

Nothing in this knowledge base keeps an append-only history of how a contract, doc, or rule reached its current form. This is a deliberate omission, not a gap: a decision log accumulates time-bound context whose relevance decays continuously, while imposing a permanent reading cost on every session that opens the file afterward. When two sources disagree — an existing doc against the real implementation, or a contract against what a person just described — the conflict is surfaced and resolved explicitly, and the resolution becomes the corrected current text. No separate record of *that* resolution is kept beyond the corrected text itself. The contract is expected to be complete and correct standing alone, because there is nothing else to fall back on.

## Durable vocabulary vs. volatile reference

Every artifact in this knowledge base is written to survive the codebase moving under it. A specific file path, filename, or function/class/variable name is never cited as proof that a convention exists — that binding breaks the moment the artifact is renamed or moved, and nothing forces the surrounding prose to notice. What is safe to name is the durable shape underneath: a pattern, a role, a module boundary, a schema or table name — the parts of a system that only change through a deliberate design decision, not a routine refactor. A delta's own contract files are the one deliberate exception, because naming the real implementation is the entire point of a contract that has to be checked against it.

This same discipline is why the knowledge base's own project-wide references stay self-contained rather than pointing at each other, and why they stay in plain, English, technical language with product or business framing stripped out even when the underlying source carried it — a docs layer that only holds engineering-relevant fact ages better than one that also has to track how a pitch or a policy phrased something.

## Non-goals

- Doesn't restate any single capability's own step-by-step mechanics, template shape, or trigger phrasing — those live with the capability itself and are expected to evolve independently of the mental model described here.
- Doesn't cover this project's product vision or scope boundaries — that belongs in a separate, dedicated reference once one exists.
- Doesn't include a fifth "execution" role for implementing a delta once its contract is written — that's ordinary coding work, not a separate delta capability.
