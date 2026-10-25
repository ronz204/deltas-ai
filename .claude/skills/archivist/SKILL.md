---
name: archivist
description: >-
  Use whenever creating or updating anything in this project's knowledge base: docs under .claude/docs/, rules under .claude/rules/, subagents under .claude/agents/, skills under .claude/skills/, the root CLAUDE.md, or a slice's spec/design file under deltas/ (<slice>.spec.md, optionally <slice>.design.md; per service in a multi-service repo). Trigger when the user asks to document architecture, write up a module, record an infrastructure or config decision, capture a coding convention, define a subagent, scaffold a skill, update CLAUDE.md, or write or amend a spec or UI-facing contract, even without saying "documentation", e.g. "we just decided X, let's write that down", "make an agent that reviews Y", "turn this into a skill", "update the CLAUDE.md", "let's put this in X's spec", "document how this screen should look and work". A spec or design that still needs information-gathering (intent, invariants, open questions) goes to the `specifier` skill first; archivist writes the result.
---

# Archivist

Creates and updates every artifact in this project's Claude Code knowledge base: context docs, rules, subagents, skills, the root CLAUDE.md, and per-slice specs/designs. What makes this different from writing any one of those from scratch is threefold — every claim is grounded in the actual codebase instead of assumption, every artifact lands in the right *kind* of file instead of getting crammed into whichever one is open, and every kind has a matching skeleton in `references/` so its shape stays consistent across instances instead of being re-derived from prose each time. This skill itself is meant to be portable: the same instructions apply whether it runs in this repo or a completely different one, so nothing below should ever hard-code this project's own stack, module names, or file paths as if they were universal facts. Live docs are under `.claude/docs/`; per-slice specs/designs live as flat files directly under `deltas/` at the project root (e.g. `deltas/<slice>.spec.md`) — no per-slice subfolder, no append-only history log. The root `deltas/` is the default; a repo genuinely split into multiple independently-deployable services instead keeps each service's own `deltas/` directory under that service's root — see `references/spec.guide.md`. This system deliberately has no history/decision-log layer: it accumulates too much time-bound context for too little ongoing value. A slice's `spec.md` (and its `design.md`, where one exists) are the only artifacts this system keeps for it — there is no third roadmap file; moving from a finished contract to code is ordinary implementation work, never a persisted knowledge-base artifact.

---

## Step 0 — Read before you write

Never write or edit any of these from memory, from what the user just said, or from what an older doc says. Read the actual source first.

- If the artifact concerns a service, module, or component, find and read the source that actually implements it — the real folder/package, its config, its existing tests. Read the whole file, not an excerpt; don't infer behavior from a filename or a related doc's description.
- If it concerns infrastructure (roles, environments, provisioned resources, hosting), check the actual provisioning config/IaC/migrations, not just what was decided in conversation — decisions drift from implementation.
- If it's a spec or design file for a slice, read that slice's current `<slice>.spec.md` (if it exists) and its `<slice>.design.md` (if present) first, so you don't contradict settled content silently — plus the actual source implementing the slice, the same discipline as any other artifact.
- **If the artifact is a new or substantially-revised `<slice>.spec.md` or `<slice>.design.md`, gate on where the content is coming from before writing.** This skill's own trigger phrases (e.g. "let's put this in X's spec") can fire even when intent, invariants/scope, or states/interactions haven't actually been established yet — don't let that self-trigger substitute for the interview. Check: is what needs to go in each section (spec: intent, scope & non-goals, invariants, deferred questions, acceptance criteria; design: scope, layout, states, interactions, consumes, deferred questions, acceptance criteria) already fully dictated in this conversation, already settled in the file being revised, or otherwise unambiguous from a single existing source? If yes, write directly — that's this skill's job. If any of it would have to be inferred from reading the implementation alone (i.e. no one has actually decided or stated it), stop and hand off to `specifier` first instead of guessing at intent or inventing an invariant. A small, obviously-settled edit (fixing one stale field, adding one already-agreed line) doesn't need the full interview — the gate is for content that isn't actually known yet, not for every touch of the file.
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
| Behavior that isn't evident from reading one component: flows across components, shared guarantees, surprising behavior and pitfalls | `.claude/docs/modules.md` — always created; the file ships with its own fill-in skeleton |
| System topology, stack choices + rationale, how parts communicate, cross-cutting patterns (auth, background jobs, etc.), and static infrastructure/provisioning facts (hosting, environments, provisioned resources) outside the persistence layer | `.claude/docs/structure.md` — always created; the file ships with its own fill-in skeleton |
| All persistence in one place: data model (entities, relationships, constraints) and database infrastructure (engine, hosting, backup/replication) together, deliberately not split against structure.md | `.claude/docs/database.md` — always created, even when the honest content is "not applicable here" (no persistent store); the file ships with its own fill-in skeleton |
| Reference explanations for non-trivial mechanisms/concepts the project depends on — how each actually works, not whether/where it's used | `.claude/docs/expertise.md` — always created, even when the honest content is "not applicable here"; the file ships with its own fill-in skeleton |
| The build sequencing philosophy: technical pillars built together deliberately, functional scope, phased roadmap, known risks to that roadmap, done criteria, stretch goals | `.claude/docs/approach.md` — always created, even when the honest content is "not applicable here" (no meaningful sequencing beyond building the thing); the file ships with its own fill-in skeleton |
| A convention that must be **applied every time** a certain kind of file is written or edited (a style rule, a naming pattern, a required check) | `.claude/rules/<topic>.md`, scoped via `paths:` frontmatter |
| A repeatable, isolated *task* — something you'd want to hand off with its own tool restrictions and its own context window | `.claude/agents/<name>.md` |
| A capability Claude Code should reach for across many tasks, potentially with bundled scripts/templates/references | `.claude/skills/<name>/SKILL.md` |
| The living, current contract for one slice/capability's logic layer — intent, contract, invariants, current state | `deltas/<slice>.spec.md` |
| The living, current contract for that same slice's UI-facing surface, where it has one — layout, states, interactions, what it consumes from spec.md's contract. Optional; only for slices with a presentation layer. | `deltas/<slice>.design.md` |

The exact doc filenames above are illustrative, not mandatory — what matters is the categorization principle (vision vs. cross-component behavior vs. topology vs. persistence) and each file staying self-contained, not the literal names. Adopt whatever breakdown a project already uses; only propose new filenames when nothing existing fits.

The doc-vs-rule line is about repetition (read once vs. applied every time). The rule-vs-agent line is about isolation: a rule injects context into whatever's already happening; an agent is a separate worker with its own system prompt, its own tool access, and its own context window, invoked for a bounded task. The agent-vs-skill line is about surface: an agent is one specialized worker; a skill is a capability (potentially with scripts/references/assets) that the main session or an agent can pull in. The spec-vs-doc line is about ownership and grain: a doc in `.claude/docs/` is the cross-project, curated reference maintained here; a slice's `spec.md` is finer-grained, lives inside the project it describes, and is expected to change as often as that slice does — a decision that stabilizes and matters beyond one slice graduates into `.claude/docs/`, it doesn't stay duplicated in both places. The spec-vs-design line (within one slice) is about layer, not shape: both are living contracts, edited in place, always describing what's true now — `spec.md` owns the logic layer (data model, business rules, invariants), `design.md` owns the presentation layer (layout, states, interactions, what it consumes). `design.md` only exists for slices with a UI-facing surface, and must reference the part of `spec.md`'s contract relevant to rendering it rather than restating an invariant or rule that already lives there — that duplication is exactly what causes the two to drift apart. `spec.md` and `design.md` are otherwise always structured as their fixed living-contract shape; the one exception is each file's own optional Context section — free-form theory or mechanism explanation that doesn't fit the rest of that file's shape without stretching it, for the rare slice that needs it. `CLAUDE.md` sits outside all of this: its job is routing to everything else, not holding durable content itself, which is why it's also the one file exempt from the self-contained/no-cross-reference rule in Step 2.

If content doesn't cleanly fit an existing file, propose either a new section in the closest existing doc or a new file following the same naming/structure pattern already in use — don't force it into an unrelated file just to avoid creating one.

**Spotting a rule candidate.** A pattern earns a `.claude/rules/<topic>.md` of its own only when all four hold: it's cross-cutting — applies to a category of files, not to one slice's behavior (if it only matters within one capability, it's a spec invariant instead, not a rule); it isn't already caught mechanically by a linter/type-checker/CI check (a rule is for what requires judgment, not what a tool already enforces); it applies every time a matching file is touched, not a one-off decision; and it would be silently violated without the reminder, not something obvious from reading the code once. This applies whether the pattern surfaces while investigating a slice (`specifier`'s own Step 1) or while reading source for any other artifact (Step 0 above) — flag it as a rule candidate rather than folding a cross-cutting convention into one slice's spec or one doc's prose, and confirm with the user before writing it: a rule fires on every future touch of a matching file, so a false positive there actively injects wrong guidance into every session afterward, costlier than a misplaced doc claim.

---

## Use the matching template

Every artifact kind above has a fill-in skeleton. Most live in `references/`: `claude.template.md` (CLAUDE.md), `docs.template.md`, `rules.template.md`, `agents.template.md`, `skills.template.md`, `spec.template.md`, `design.template.md`. The six fixed `.claude/docs/*.md` categories — `overview.md`, `modules.md`, `structure.md`, `database.md`, `expertise.md`, `approach.md` — are the one exception: each ships as part of this harness carrying its own fill-in skeleton directly in the file, since it's the same file that ends up living in the target project rather than a separate copy meant to be re-derived from. If one of those six still holds only its guidance-comment skeleton (unfilled), that skeleton is what gets filled in place; `docs.template.md` stays the fallback shape only for a genuinely new one-off doc category with no shipped skeleton of its own. Read the matching skeleton before writing, and fill it in rather than re-deriving the shape from the prose in the steps below each time — the skeletons carry the structure, the steps below carry the reasoning behind that structure.

**`spec.template.md` and `design.template.md` each carry one deliberate exception to their own fixed shape: an optional Context section.** It's meant to stay a loose, improvised write-up shaped by whatever the slice actually needs (a table here, a diagram there, a couple of paragraphs of explanation) rather than forcing supplementary theory into the same sections as the rest of the contract. Still ground it in the real source/technology per Step 0, and still keep it out of decision-log territory (see `references/spec.guide.md`) — just don't reach for a fixed skeleton to structure it, and leave it out entirely for the common case where a slice doesn't need it.

---

## Step 2 — Writing conventions for docs and rules

These apply to `.claude/docs/*.md` and `.claude/rules/*.md`. They exist because earlier drafts got corrected for violating them. The volatile-vs-durable reference rule below also governs `.claude/skills/*` and `.claude/agents/*` — see `references/agents.guide.md` and `references/skills.guide.md`:

- **No references to volatile implementation artifacts.** Never cite a specific file path, filename, or function/class/variable name as proof a convention exists, and never copy a table that already lives in a config file (e.g. a path-alias list belongs in the project's own config file alone — point at it, don't duplicate it). These break the moment the artifact moves or gets renamed, and nothing forces the doc to notice. Describe the *pattern*, not *where it currently lives*. Exception: a single-instance, tool-mandated config filename (`package.json`, `tsconfig.json`, `pyproject.toml`, `Dockerfile`, `compose.yml`) is safe to name — renaming one breaks the tool itself, so it isn't actually volatile.
- **`.claude/docs/*.md` gets one carve-out from the rule above: durable architecture vocabulary.** A schema/table name, a named infrastructure role, a module boundary, a security invariant — these describe the system's actual shape and only change via a deliberate architecture decision, which is exactly what a docs file exists to capture. What still doesn't belong, even in docs, is *where in the source tree* that shape is implemented — a specific file path stays off-limits there too. `.claude/rules/*.md` and `.claude/skills/*` don't get this carve-out — they describe process/convention, not the system's architecture, so they stay fully agnostic to implementation nouns as well as locations.
- **Write only what the code can't tell.** Before a statement goes into a doc, ask whether a session could get it from the code or the tool config in a few reads (a component's behavior, a schema, a dependency list, a directory map). If so, leave it out, or point at where it lives in general terms. Keep what the code doesn't show: the reason behind a decision, a constraint or a pitfall, a convention no tool enforces, a fact that lives outside the repo (hosting, environments, roadmap, vision). A doc with nothing like that says "not applicable here" in one line instead of being filled — content copied from the code drifts and distracts.
- **English**, regardless of what the source material or conversation was in.
- **Pure technical POV.** Strip product pitch, ROI framing, competitive positioning, sales language, business/legal framing — even if the source had it. If a business fact drives an engineering decision (e.g. a data-protection law shaping a storage choice), state the *engineering implication*, not the business rationale.
- **Self-contained files.** Docs in `.claude/docs/` don't reference each other ("see X.md") — each should stand alone. The one exception is a doc split to fit its context budget: the entry doc points to its own detail files (Step 8), and those detail files still stand alone. Rules may reference a doc once if the split genuinely creates a dependency — keep it to the minimum, prefer restating a short fact over adding a pointer. `CLAUDE.md` is the deliberate exception on the other end — its entire job is pointing to everything else, so cross-references there are expected, not a violation.
- **Plain section headers**, no emoji.
- **Tables** for comparisons, stack choices with rationale, risk/complexity summaries.
- **Fenced code blocks** for SQL, JSON, config snippets, ASCII diagrams — never describe a schema or query in prose when a code block says it exactly.
- **Explain why, not just what.** A convention stated without its reasoning gets silently violated the first time someone doesn't see why it matters.
- **Length:** every artifact has a ceiling set by how often it loads, kept in the project's context-budget config (Step 8) — a doc growing past its ceiling is a sign it needs splitting, and a rule stays scoped to one concern.
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

## Steps 3–6 — Type-specific guides

Conventions that apply to only one kind of artifact live in a guide beside its template, so they load only when that kind is being written. Before writing, read the guide for the kind you're writing:

| Writing | Read first |
|---|---|
| A subagent (`.claude/agents/`) | `references/agents.guide.md` |
| A skill (`.claude/skills/<name>/SKILL.md`) | `references/skills.guide.md` |
| A spec or design (`deltas/`) | `references/spec.guide.md` |
| `CLAUDE.md` | `references/claude.guide.md` |

Docs and rules need no extra guide: Step 2 covers them.

---

## Step 7 — Before presenting the draft

- Re-read the finished artifact once as if you'd never seen the project: does every claim trace back to something actually read in Step 0, not something assumed?
- **For a doc, apply the filter from Step 2:** strike every statement a session could get from the code or tool config in a few reads, and keep only what the code doesn't show.
- **Check the reverse direction too, whenever the artifact distills an existing source document** (a root-level vision/approach file feeding `overview.md`/`approach.md`, or any case where one file is being reorganized into another): does every notable section or decision in that source have a home somewhere in what was written, or a deliberate, stated reason it was left out? A skeleton's fixed sections make it easy to map only the content that has an obvious slot and silently drop what doesn't — that's exactly as much a correctness failure as inventing a claim that isn't there, and it's invisible to the check above since the omitted content genuinely was read in Step 0, just never placed anywhere.
- **Scan for volatile references** (Step 2): any file path, filename, or function/class/variable name cited as proof — outside `<slice>.spec.md`/`<slice>.design.md`, where that's the point. In `.claude/docs/*.md`, confirm what survived is architecture vocabulary (schema/role/module/invariant), not a source-tree location.
- Check it against the length/format norms for its type (doc, rule, agent, skill, spec, design, CLAUDE.md) and confirm it matches its `references/*.template.md` skeleton — except a spec/design's optional Context section, which has no fixed skeleton by design; for that section, confirm instead that it stayed free-form and didn't drift into decision-log shape.
- If you touched an existing file, confirm you didn't reintroduce a cross-reference between docs, or product/business language that was previously stripped — unless the file is `CLAUDE.md`, where cross-references are the point.
- If you wrote or edited a `<slice>.spec.md`, confirm the file itself reads as complete and current — there's no history entry to fall back on, so a claim left implicit or half-updated stays that way until the next edit. Also confirm every invariant and acceptance criterion carries a verification tag, and that each `test` tag's file and name were confirmed to exist. Also confirm every new or revised invariant is one claim in an EARS form with an observable response, and every acceptance criterion is a Given/When/Then line.
- If you wrote or edited a `<slice>.design.md`, apply the same completeness check, plus confirm it didn't restate an invariant or business rule that already lives in `<slice>.spec.md` — it should reference that contract, not duplicate it.
- If Step 0 turned up a conflict between artifacts, or between an artifact and the code, lead with that when presenting the draft — don't bury it at the end.
---

## Step 8 — Check the context budget

Every artifact has a ceiling on how much context it may cost, set by how often it loads — see the context budget rule. After writing or editing anything in the knowledge base, run `sh .claude/skills/archivist/codebase/budget.sh` (read-only, costs no model tokens) and read only its "over a ceiling" and "broken pointers" sections.

- If what you touched is over its ceiling, don't trim it to fit and don't ship it over: fit it by splitting. Read `references/split.guide.md`, show the user the split plan (what stays, what moves where, which pointer replaces it), and apply it only once they confirm.
- If a pointer is reported broken, fix it before presenting.
- If `sh` isn't available here, compare the file's line count against the ceilings in `.claude/budget.conf` and say the check was approximate.
