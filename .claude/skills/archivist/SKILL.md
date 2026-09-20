---
name: archivist
description: >-
  Use whenever creating or updating anything in this project's Claude Code knowledge base — docs under .claude/docs/, rules under .claude/rules/, subagents under .claude/agents/, skills under .claude/skills/, the root orientation file CLAUDE.md, or a slice's spec/design file under deltas/ at the project root (or, in a repo genuinely split into multiple services, under a service's own deltas/ directory) (flat files — <slice>.spec.md, optionally <slice>.design.md). Trigger this any time the user asks to document architecture, write up a module, record an infrastructure/config decision, capture a coding convention, define a new subagent, scaffold a new skill, update CLAUDE.md, write or amend a <slice>.spec.md, or write/amend a slice's UI-facing contract — even without saying "documentation", e.g. "we just decided X, let's write that down", "make an agent that reviews Y", "turn this into a skill", "actualiza el CLAUDE.md", "escribamos esto en el spec de X", "documenta cómo debe verse y funcionar esta pantalla". Always inspect the actual codebase before writing anything; never document from memory or from the conversation alone. For a spec (or design) that needs real information-gathering first (intent, invariants, open questions the code alone can't answer), that's the `specifier` skill's job — archivist writes the result into the correct shape, it doesn't run the interview.
---

# Archivist

Creates and updates every artifact in this project's Claude Code knowledge base: context docs, rules, subagents, skills, the root CLAUDE.md, and per-slice specs/designs. What makes this different from writing any one of those from scratch is threefold — every claim is grounded in the actual codebase instead of assumption, every artifact lands in the right *kind* of file instead of getting crammed into whichever one is open, and every kind has a matching skeleton in `references/` so its shape stays consistent across instances instead of being re-derived from prose each time. This skill itself is meant to be portable: the same instructions apply whether it runs in this repo or a completely different one, so nothing below should ever hard-code this project's own stack, module names, or file paths as if they were universal facts. Live docs are under `.claude/docs/`; per-slice specs/designs live as flat files directly under `deltas/` at the project root (e.g. `deltas/<slice>.spec.md`) — no per-slice subfolder, no append-only history log. The root `deltas/` is the default; a repo genuinely split into multiple independently-deployable services instead keeps each service's own `deltas/` directory under that service's root — see Step 5. This system deliberately has no history/decision-log layer: it accumulates too much time-bound context for too little ongoing value. A slice's `spec.md` (and its `design.md`, where one exists) are the only artifacts this system keeps for it — there is no third roadmap file; moving from a finished contract to code is ordinary implementation work, never a persisted knowledge-base artifact.

---

## Step 0 — Read before you write

Never write or edit any of these from memory, from what the user just said, or from what an older doc says. Read the actual source first.

- If the artifact concerns a service, module, or component, find and read the source that actually implements it — the real folder/package, its config, its existing tests. Read the whole file, not an excerpt; don't infer behavior from a filename or a related doc's description.
- If it concerns infrastructure (roles, environments, provisioned resources, hosting), check the actual provisioning config/IaC/migrations, not just what was decided in conversation — decisions drift from implementation.
- If it's a spec or design file for a slice, read that slice's current `<slice>.spec.md` (if it exists) and its `<slice>.design.md` (if present) first, so you don't contradict settled content silently — plus the actual source implementing the slice, the same discipline as any other artifact.
- **If the artifact is a new or substantially-revised `<slice>.spec.md` or `<slice>.design.md`, gate on where the content is coming from before writing.** This skill's own trigger phrases (e.g. "escribamos esto en el spec de X") can fire even when intent, invariants/scope, or states/interactions haven't actually been established yet — don't let that self-trigger substitute for the interview. Check: is what needs to go in each section (spec: intent, scope & non-goals, invariants, deferred questions, acceptance criteria; design: scope, layout, states, interactions, consumes, deferred questions, acceptance criteria) already fully dictated in this conversation, already settled in the file being revised, or otherwise unambiguous from a single existing source? If yes, write directly — that's this skill's job. If any of it would have to be inferred from reading the implementation alone (i.e. no one has actually decided or stated it), stop and hand off to `specifier` first instead of guessing at intent or inventing an invariant. A small, obviously-settled edit (fixing one stale field, adding one already-agreed line) doesn't need the full interview — the gate is for content that isn't actually known yet, not for every touch of the file.
- If the user references a decision made earlier in chat ("ya cambiamos a X"), verify it landed in code before documenting it as current. If it hasn't landed yet, say so and ask whether to document it as current or as planned.
- **Never cite a concrete path, folder, or filename that isn't actually confirmed to exist or be settled.** This applies everywhere a path can appear — a rule's `paths:` frontmatter glob, an example in a doc — not just the volatile-reference wording in Step 2. Before writing one in, check with a real directory listing that it exists, or confirm it's already a settled decision recorded elsewhere (e.g. a service name in `CLAUDE.md`'s repo layout table). If it's neither, don't invent a plausible-looking one: scope to the nearest boundary that *is* settled, or ask the user. An unconfirmed folder name in a rule's `paths:` is worse than a vague one — the rule silently never fires until a file happens to match the guessed layout, and nothing flags that it's dead.
- Skim what already exists in `.claude/docs/`, `.claude/rules/`, `.claude/agents/`, `.claude/skills/`, and `CLAUDE.md` so you don't duplicate or contradict something that's already documented elsewhere.
- While reading source for any artifact, note any transversal, not-tool-enforced pattern that looks like it should apply every time a matching file is touched — that's a rule candidate (see "Spotting a rule candidate" below), not content for the artifact currently being written. Confirm with the user it's intentional before creating it as its own `.claude/rules/<topic>.md`.

If code contradicts an existing doc, or contradicts what the user just described, don't silently resolve it — surface the conflict and let the user decide which one is current.

---

## Step 1 — Figure out what kind of artifact this is

| Content is... | Goes in |
|---|---|
| What this project is, how to get oriented fast, where the rest of the knowledge base lives | `CLAUDE.md` |
| Product/system vision, scope boundaries, what problem the system solves, the purpose of each major part at a conceptual level | `.claude/docs/overview.md` — always created; the file ships with its own fill-in skeleton |
| Per-component functional reference: flows, data shape, what a component actually does step by step | `.claude/docs/modules.md` — always created; the file ships with its own fill-in skeleton |
| System topology, stack choices + rationale, how parts communicate, cross-cutting patterns (auth, background jobs, etc.), and static infrastructure/provisioning facts (hosting, environments, provisioned resources) outside the persistence layer | `.claude/docs/structure.md` — always created; the file ships with its own fill-in skeleton |
| All persistence in one place: data model (entities, relationships, constraints) and database infrastructure (engine, hosting, backup/replication) together, deliberately not split against structure.md | `.claude/docs/database.md` — always created, even when the honest content is "not applicable here" (no persistent store); the file ships with its own fill-in skeleton |
| Reference explanations for non-trivial mechanisms/concepts the project depends on — how each actually works, not whether/where it's used | `.claude/docs/expertise.md` — always created, even when the honest content is "not applicable here"; the file ships with its own fill-in skeleton |
| The build sequencing philosophy: technical pillars built together deliberately, functional scope, phased roadmap, known risks to that roadmap, done criteria, stretch goals | `.claude/docs/approach.md` — always created, even when the honest content is "not applicable here" (no meaningful sequencing beyond building the thing); the file ships with its own fill-in skeleton |
| A convention that must be **applied every time** a certain kind of file is written or edited (a style rule, a naming pattern, a required check) | `.claude/rules/<topic>.md`, scoped via `paths:` frontmatter |
| A repeatable, isolated *task* — something you'd want to hand off with its own tool restrictions and its own context window | `.claude/agents/<name>.md` |
| A capability Claude Code should reach for across many tasks, potentially with bundled scripts/templates/references | `.claude/skills/<name>/SKILL.md` |
| The living, current contract for one slice/capability's logic layer — intent, contract, invariants, current state | `deltas/<slice>.spec.md` |
| The living, current contract for that same slice's UI-facing surface, where it has one — layout, states, interactions, what it consumes from spec.md's contract. Optional; only for slices with a presentation layer. | `deltas/<slice>.design.md` |

The exact doc filenames above are illustrative, not mandatory — what matters is the categorization principle (vision vs. functional reference vs. topology vs. persistence) and each file staying self-contained, not the literal names. Adopt whatever breakdown a project already uses; only propose new filenames when nothing existing fits.

The doc-vs-rule line is about repetition (read once vs. applied every time). The rule-vs-agent line is about isolation: a rule injects context into whatever's already happening; an agent is a separate worker with its own system prompt, its own tool access, and its own context window, invoked for a bounded task. The agent-vs-skill line is about surface: an agent is one specialized worker; a skill is a capability (potentially with scripts/references/assets) that the main session or an agent can pull in. The spec-vs-doc line is about ownership and grain: a doc in `.claude/docs/` is the cross-project, curated reference maintained here; a slice's `spec.md` is finer-grained, lives inside the project it describes, and is expected to change as often as that slice does — a decision that stabilizes and matters beyond one slice graduates into `.claude/docs/`, it doesn't stay duplicated in both places. The spec-vs-design line (within one slice) is about layer, not shape: both are living contracts, edited in place, always describing what's true now — `spec.md` owns the logic layer (data model, business rules, invariants), `design.md` owns the presentation layer (layout, states, interactions, what it consumes). `design.md` only exists for slices with a UI-facing surface, and must reference the part of `spec.md`'s contract relevant to rendering it rather than restating an invariant or rule that already lives there — that duplication is exactly what causes the two to drift apart. `spec.md` and `design.md` are otherwise always structured as their fixed living-contract shape; the one exception is each file's own optional Context section — free-form theory or mechanism explanation that doesn't fit the rest of that file's shape without stretching it, for the rare slice that needs it. `CLAUDE.md` sits outside all of this: its job is routing to everything else, not holding durable content itself, which is why it's also the one file exempt from the self-contained/no-cross-reference rule in Step 2.

If content doesn't cleanly fit an existing file, propose either a new section in the closest existing doc or a new file following the same naming/structure pattern already in use — don't force it into an unrelated file just to avoid creating one.

**Spotting a rule candidate.** A pattern earns a `.claude/rules/<topic>.md` of its own only when all four hold: it's cross-cutting — applies to a category of files, not to one slice's behavior (if it only matters within one capability, it's a spec invariant instead, not a rule); it isn't already caught mechanically by a linter/type-checker/CI check (a rule is for what requires judgment, not what a tool already enforces); it applies every time a matching file is touched, not a one-off decision; and it would be silently violated without the reminder, not something obvious from reading the code once. This applies whether the pattern surfaces while investigating a slice (`specifier`'s own Step 1) or while reading source for any other artifact (Step 0 above) — flag it as a rule candidate rather than folding a cross-cutting convention into one slice's spec or one doc's prose, and confirm with the user before writing it: a rule fires on every future touch of a matching file, so a false positive there actively injects wrong guidance into every session afterward, costlier than a misplaced doc claim.

---

## Use the matching template

Every artifact kind above has a fill-in skeleton. Most live in `references/`: `claude.template.md` (CLAUDE.md), `docs.template.md`, `rules.template.md`, `agents.template.md`, `skills.template.md`, `spec.template.md`, `design.template.md`. The six fixed `.claude/docs/*.md` categories — `overview.md`, `modules.md`, `structure.md`, `database.md`, `expertise.md`, `approach.md` — are the one exception: each ships as part of this harness carrying its own fill-in skeleton directly in the file, since it's the same file that ends up living in the target project rather than a separate copy meant to be re-derived from. If one of those six still holds only its guidance-comment skeleton (unfilled), that skeleton is what gets filled in place; `docs.template.md` stays the fallback shape only for a genuinely new one-off doc category with no shipped skeleton of its own. Read the matching skeleton before writing, and fill it in rather than re-deriving the shape from the prose in the steps below each time — the skeletons carry the structure, the steps below carry the reasoning behind that structure.

**`spec.template.md` and `design.template.md` each carry one deliberate exception to their own fixed shape: an optional Context section.** It's meant to stay a loose, improvised write-up shaped by whatever the slice actually needs (a table here, a diagram there, a couple of paragraphs of explanation) rather than forcing supplementary theory into the same sections as the rest of the contract. Still ground it in the real source/technology per Step 0, and still keep it out of decision-log territory (see Step 5) — just don't reach for a fixed skeleton to structure it, and leave it out entirely for the common case where a slice doesn't need it.

---

## Step 2 — Writing conventions for docs and rules

These apply to `.claude/docs/*.md` and `.claude/rules/*.md`. They exist because earlier drafts got corrected for violating them. The volatile-vs-durable reference rule below also governs `.claude/skills/*` and `.claude/agents/*` — see Steps 3–4:

- **No references to volatile implementation artifacts.** Never cite a specific file path, filename, or function/class/variable name as proof a convention exists, and never copy a table that already lives in a config file (e.g. a path-alias list belongs in the project's own config file alone — point at it, don't duplicate it). These break the moment the artifact moves or gets renamed, and nothing forces the doc to notice. Describe the *pattern*, not *where it currently lives*. Exception: a single-instance, tool-mandated config filename (`package.json`, `tsconfig.json`, `pyproject.toml`, `Dockerfile`, `compose.yml`) is safe to name — renaming one breaks the tool itself, so it isn't actually volatile.
- **`.claude/docs/*.md` gets one carve-out from the rule above: durable architecture vocabulary.** A schema/table name, a named infrastructure role, a module boundary, a security invariant — these describe the system's actual shape and only change via a deliberate architecture decision, which is exactly what a docs file exists to capture. What still doesn't belong, even in docs, is *where in the source tree* that shape is implemented — a specific file path stays off-limits there too. `.claude/rules/*.md` and `.claude/skills/*` don't get this carve-out — they describe process/convention, not the system's architecture, so they stay fully agnostic to implementation nouns as well as locations.
- **English**, regardless of what the source material or conversation was in.
- **Pure technical POV.** Strip product pitch, ROI framing, competitive positioning, sales language, business/legal framing — even if the source had it. If a business fact drives an engineering decision (e.g. a data-protection law shaping a storage choice), state the *engineering implication*, not the business rationale.
- **Self-contained files.** Docs in `.claude/docs/` don't reference each other ("see X.md") — each should stand alone. Rules may reference a doc once if the split genuinely creates a dependency — keep it to the minimum, prefer restating a short fact over adding a pointer. `CLAUDE.md` is the deliberate exception on the other end — its entire job is pointing to everything else, so cross-references there are expected, not a violation.
- **Plain section headers**, no emoji.
- **Tables** for comparisons, stack choices with rationale, risk/complexity summaries.
- **Fenced code blocks** for SQL, JSON, config snippets, ASCII diagrams — never describe a schema or query in prose when a code block says it exactly.
- **Explain why, not just what.** A convention stated without its reasoning gets silently violated the first time someone doesn't see why it matters.
- **Length:** docs stay ~100–200 lines — a section growing past that is a sign it belongs in a more specific file. Rules can run longer since they only load conditionally, but stay scoped to one concern.
- **Non-goals sections earn their place** when a scope boundary is easy to violate by accident.

### Rule mechanics

`.claude/rules/<name>.md` loads automatically when Claude Code reads a file matching the `paths:` glob patterns in the frontmatter — same priority as `CLAUDE.md`.

```yaml
---
paths:
  - "<glob pattern, quoted — YAML requires it for patterns starting with * or {>"
---
```

- Quote every glob pattern (YAML requires it for patterns starting with `*` or `{`).
- Omit `paths:` entirely only for a rule meant to apply globally — this should be rare.
- One rule file per concern — a new convention with different trigger paths is a new rule, not an addition to an unrelated one.

---

## Step 3 — Writing subagents (`.claude/agents/`)

A subagent is a markdown file with YAML frontmatter; the body is its system prompt. It runs in its own context window with its own tool access — use it for a bounded, repeatable task you'd otherwise delegate the same way every time.

```yaml
---
name: spec-invariant-checker
description: Reviews changes to a slice's implementation against the invariants declared in its <slice>.spec.md before merge. Use after editing source files inside a slice that has a spec.
tools: Read, Glob, Grep
model: sonnet
---

You are checking a diff against the invariants declared in <slice>.spec.md for the slice it touches. For each invariant listed:
- confirm the changed code still upholds it
- if an invariant is no longer upheld, say so explicitly — don't decide on the user's behalf whether that's acceptable
- ...
```

- **`description` is the trigger** — same rule as skills: be specific about when to invoke it, don't undersell it.
- **Restrict `tools` to the minimum** the task needs — a review/analysis agent shouldn't get write access it doesn't use.
- **Choose `model` deliberately**: cheaper/faster models for narrow search or lint-style checks, the default for anything requiring judgment.
- **Ground the system prompt in this project's actual conventions** rather than generic reviewer advice — pull specifics from `.claude/rules/` and `.claude/docs/` rather than restating boilerplate. Same volatile-vs-durable reference rule as Step 2: describe the pattern the agent should check for, never cite a specific file as the example of it.
- Names must be unique across `.claude/agents/` — a collision gets silently discarded, not merged.
- Use `references/agents.template.md` as the fill-in skeleton; the example above shows the shape it produces.

---

## Step 4 — Writing new skills (`.claude/skills/<name>/SKILL.md`)

A skill's `name` + `description` are always in context; the body loads only when it triggers; anything under `scripts/`, `references/`, or `assets/` loads on demand.

- **`description` does double duty**: state what the skill does *and* when to use it, and lean slightly pushy — Claude tends to under-trigger skills, so spell out phrasings and contexts explicitly rather than trusting a vague description to be inferred.
- Keep the `SKILL.md` body under ~500 lines. If it's growing past that, split stable reference material into `references/*.md` and point to it from the body rather than inlining everything.
- Write instructions in the imperative, and explain *why* rather than issuing bare musts — same principle as docs and rules.
- Every new skill goes through the same Step 0 discipline: don't invent its instructions from a generic template — `references/skills.template.md` gives the *shape*, but the actual content still has to come from how the task is really done in this project, checking existing docs/rules/agents first so the new skill doesn't duplicate or contradict what already exists.
- Same volatile-vs-durable reference rule as Step 2 applies to `SKILL.md` and everything under its `references/` — a skill's instructions and reference material should point at a durable structure (a project, a directory holding one concern) or describe a pattern in the abstract, never cite a specific implementation file as the proof that pattern exists in this repo.

---

## Step 5 — Writing specs and designs (`deltas/`)

A slice is a cohesive capability (e.g. a `checkout-flow` slice), not one source file — pick the grain the same way a bounded context would be picked, not by file structure. Slices live as flat files directly under `deltas/` at the project root — `<slice>.spec.md`, optionally `<slice>.design.md` — never a per-slice subfolder. This root `deltas/` is the default and covers the common case: one project, one place for its slices. Favor fewer, right-sized slices over many thin ones: if two slices' spec.md keep changing together for the same reason, that's usually a sign they're one bounded context, not two — merging them costs less upkeep than maintaining the pair in lockstep.

**The exception is a repo genuinely split into multiple independently-deployable services (e.g. a backend and a separate frontend, each with its own root)** — not a project that merely has multiple top-level directories, but one where each is really its own deployable unit. There, `deltas/` moves down one level, into each service's own root, and a slice's siblings don't have to live under the same service's `deltas/`: each file lives under whichever service actually owns that layer — `<slice>.design.md` under the service rendering the UI, `<slice>.spec.md` under whichever service implements the logic it describes — same slice name, different `deltas/` directories, if those are two different services. Don't default to co-locating them just because they share a slice name, and don't reach for this structure just because a project happens to have a `frontend/`/`backend/`-shaped layout — confirm the split is real before moving off the root default. Since the files are no longer next to each other, each must name the other service(s) it pairs with (a service name is durable architecture vocabulary here, not a volatile reference) rather than assuming a reader has every sibling open at once — this applies in both directions: `design.md` names which service's `spec.md` it renders, and that `spec.md` should in turn note where its `design.md` (and any sibling `spec.md`) lives, so either file is a complete map of the slice on its own.

**A slice can also have more than one `<slice>.spec.md` — one per service, each scoped to the logic *that* service owns.** This isn't duplication: a backend's `spec.md` might own session validity and data rules while a frontend's own `spec.md` for the same slice name owns that service's own logic (e.g. client-side navigation/routing behavior) — genuinely different content, deliberately sharing a slice name because they're the same capability seen from two codebases. Each such `spec.md` must state plainly what it owns versus what it only reacts to from its sibling(s), so a reader of either one knows immediately which decisions are made there and which are made elsewhere.

- **`<slice>.spec.md` is living, current truth for the logic layer** — edited in place as the slice's contract changes. It should always describe *what is true now*, not a history of how it got there. There is no append-only decision log backing it up — this file has to be complete and correct on its own.
- **`<slice>.design.md` is living, current truth for the presentation layer, and optional** — create it only for a slice that actually has a UI-facing surface; a purely backend slice never needs one. Same discipline as `spec.md` (edited in place, always current, no history log), scoped to layout, states, interactions, and what it consumes. It must not restate an invariant or business rule that already lives in `spec.md` — reference the relevant part of that contract instead, so the two files can't silently drift apart on a shared concern.
- **Either file's optional Context section is the one free-form exception** — supplementary theory that doesn't fit the rest of that file's contract shape (domain background, how a dependency's mechanism actually works, why a rule behaves the way it does). Create it only when the slice genuinely needs it — most never do. It is explicitly **not** a history replacement: never write it as a decision log ("we changed X because Y", superseded-decision narratives, timestamps) — that kind of time-bound content is deliberately not kept in this system. If content is decision-shaped, fold the current state into the rest of `spec.md` or `design.md` instead, or leave it out.
- **Reconcile conflicts explicitly, don't pick a side quietly.** If writing a spec or design surfaces a mismatch between the real implementation and an existing `.claude/docs/*` claim, fix the doc in the direction the code/decision actually supports, and update `spec.md`/`design.md` in place to match — no separate record of the resolution is kept beyond the corrected text itself.
- **Retiring a slice is deletion, not archival.** If a slice's implementation is confirmed intentionally removed — via a `sentinel` `Gone` verdict, or the user directly confirming it — delete its `<slice>.spec.md` and `<slice>.design.md` outright rather than moving them to an archive location or marking them retired in place. This system keeps no history layer; git history is already the record of what the slice's contract used to say.
- **Specs (and designs) are frequently handed off from the `specifier` skill**, which does the information-gathering (intent, invariants, open questions — the parts code alone can't answer) and the user Q&A. Treat that handoff as vetted input, but still verify it against the actual code before writing, per Step 0 — `specifier` investigates and asks, `archivist` is still the one that writes. This is the normal path for anything nontrivial; writing a spec/design directly (no `specifier` handoff) is only valid when Step 0's gate confirms the content is already fully known, not when this skill's own trigger phrases fired first.

---

## Step 6 — Writing CLAUDE.md

`CLAUDE.md` is the one artifact this system loads into *every* session automatically, so it behaves differently from everything else it routes to:

- **It's a router, not a reference.** Its job is to help a session get oriented fast and find the right deeper file — not to hold the actual content of what it points to. Anything that would make it a second copy of a doc belongs in the doc instead, with a pointer left behind.
- **Cross-references are expected here**, unlike `.claude/docs/*.md` — see Step 2.
- **Keep it short.** Every line costs context on every single turn, regardless of whether that turn needs it — this is the strongest length pressure of any artifact in this system.
- **Locate the existing file before assuming a path.** Different projects keep it in different places (repo root, or under `.claude/`) — check before creating a second one elsewhere.
- Use `references/claude.template.md`: short project description, knowledge base layout, repo layout, setup/common commands, a pointer-not-copy summary of `.claude/settings.json`, and any project-wide convention that isn't already its own rule file.
- Same Step 0 discipline applies to every factual claim in it — pull commands from the actual package manifest/task runner/CI config, pull repo layout from the actual tree, never from what a similar project usually looks like.

---

## Step 7 — Before presenting the draft

- Re-read the finished artifact once as if you'd never seen the project: does every claim trace back to something actually read in Step 0, not something assumed?
- **Check the reverse direction too, whenever the artifact distills an existing source document** (a root-level vision/approach file feeding `overview.md`/`approach.md`, or any case where one file is being reorganized into another): does every notable section or decision in that source have a home somewhere in what was written, or a deliberate, stated reason it was left out? A skeleton's fixed sections make it easy to map only the content that has an obvious slot and silently drop what doesn't — that's exactly as much a correctness failure as inventing a claim that isn't there, and it's invisible to the check above since the omitted content genuinely was read in Step 0, just never placed anywhere.
- **Scan for volatile references** (Step 2): any file path, filename, or function/class/variable name cited as proof — outside `<slice>.spec.md`/`<slice>.design.md`, where that's the point. In `.claude/docs/*.md`, confirm what survived is architecture vocabulary (schema/role/module/invariant), not a source-tree location.
- Check it against the length/format norms for its type (doc, rule, agent, skill, spec, design, CLAUDE.md) and confirm it matches its `references/*.template.md` skeleton — except a spec/design's optional Context section, which has no fixed skeleton by design; for that section, confirm instead that it stayed free-form and didn't drift into decision-log shape.
- If you touched an existing file, confirm you didn't reintroduce a cross-reference between docs, or product/business language that was previously stripped — unless the file is `CLAUDE.md`, where cross-references are the point.
- If you wrote or edited a `<slice>.spec.md`, confirm the file itself reads as complete and current — there's no history entry to fall back on, so a claim left implicit or half-updated stays that way until the next edit.
- If you wrote or edited a `<slice>.design.md`, apply the same completeness check, plus confirm it didn't restate an invariant or business rule that already lives in `<slice>.spec.md` — it should reference that contract, not duplicate it.
- If Step 0 turned up a conflict between artifacts, or between an artifact and the code, lead with that when presenting the draft — don't bury it at the end.