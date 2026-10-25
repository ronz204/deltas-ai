# NEXT — Candidate directions for the next version of Delta

A working document, not a contract. It records the research and brainstorm done on 2026-10-25 so the next version of Delta can be chosen from it later. Nothing here is committed scope. Items are ordered by suggested priority, not by decision.

Guiding constraints for any item below: Delta must stay **robust** (verifiable, not vibes), **portable** (no lock-in to one agent or runtime), and **result-producing** (it must measurably help an agent, not just look tidy).

---

## 1. Where Delta stands today

- Five single-responsibility skills: Surveyor (bootstrap), Specifier (interview), Archivist (sole writer of the knowledge base), Sentinel (verify contract vs. code) and the optional Executor (build code against a contract).
- Living documentation: files describe what is true now, edited in place, no append-only decision log (git is the history).
- Three layers: `CLAUDE.md` router, six fixed `.claude/docs/*` files, per-feature `<slice>.spec.md` / `<slice>.design.md` under `deltas/`.
- Packaged for Claude Code only (`.claude/`); Cursor support was removed on 2026-10-25. No runtime, no CLI: markdown and skills, plus one optional read-only shell script.
- Writes application code only through the optional Executor, from a plan you approve and that is derived from a spec. Does not keep a roadmap or task list per slice.

## 2. Comparison with Spec Kit

| | Spec Kit (GitHub) | Delta |
|---|---|---|
| Philosophy | Spec-driven development: the spec is the starting point to **generate code** | Living documentation: the spec is a **contract of what is true now** |
| Flow | Linear per feature: constitution, specify, clarify, plan, tasks, analyze, implement | Five roles: surveyor, specifier, archivist, sentinel, executor |
| Writes code | Yes (`/implement`) | Only through the optional Executor; its plan is derived from unverified claims, never stored |
| Artifacts per feature | `spec.md`, `plan.md`, `tasks.md`, `research.md`, `data-model.md`, `contracts/`; they accumulate | One `<slice>.spec.md` (+ `design.md` if UI), edited in place |
| Lifecycle | A spec becomes historical once the feature ships | A contract stays current while the capability exists |
| Verification | `/analyze` checks spec/plan/tasks consistency **before** implementing | Sentinel checks contract vs. real code **after** implementing |
| Brownfield | Aimed mostly at new features | Surveyor bootstraps an existing repo |
| Decision history | Accumulates in artifacts | None by design |
| Agent support | Many agents via a CLI | Claude Code via a copied folder |

**Where Delta wins:** drift detection over time, a KB that does not rot, lower context cost, brownfield onboarding, slices scoped by capability not by layer, zero runtime dependency.

**Where Delta loses:** a spec-to-code path only through the optional Executor (no stored plan or tasks), less maturity and community, every write funnels through Archivist (friction), Sentinel audits only specs/designs, no executable "constitution" check before building.

**Positioning:** Spec Kit answers "how do I build this feature from a spec"; Delta answers "how do I keep project context true and verifiable". They compose: build with Spec Kit (or a plain plan), keep the resulting contract honest with Delta.

## 3. Research findings (external)

Sources are largely vendor-adjacent blogs and secondary summaries; treat specifics as indicative and verify primaries before relying on any number.

1. **The closest competitor is OpenSpec, not Spec Kit.** It has a delta format (ADDED / MODIFIED / REMOVED), is lightweight, and targets brownfield changes — Delta's own territory. Reviewers say its specs can go static during implementation, which is exactly what Sentinel addresses. One head-to-head reported OpenSpec cheaper and faster with fewer spec lines, while Spec Kit passed slightly more acceptance checks (13/13 vs 12/13).
2. **Spec Kit is heavy for small changes.** Four artifacts per feature; comparisons recommend a workaround for small modifications. Delta's one-spec-per-slice is cheaper.
3. **Other tools:** Kiro (AWS, closed IDE, EARS notation requirements/design/tasks) and BMAD (heavy multi-agent simulated agile team, high token cost). Consensus: no single tool wins; fit depends on context.
4. **Context files can hurt.** The ETH Zurich paper "Evaluating AGENTS.md" (arXiv 2602.11988; 138-task Python benchmark; tested Claude Code, Codex, Qwen Code) found LLM-generated context files tend to **reduce task success** and raise inference cost by over 20%; developer-written ones give at best modest gains. Proposed cause: broad architectural overviews push agents into unbounded exploration. **This is a direct risk to Surveyor (generated docs) and to the six always-loaded docs.**
5. **Skills vs. AGENTS.md is contested.** A secondhand summary of a Vercel eval claims a compressed AGENTS.md docs index (100%) beat skills (79%), partly because the agent sometimes did not load the relevant skill. Primary source not found; narrow to their tasks and setup. Lesson: run our own evals.
6. **Automated drift detection is nearly absent.** Sources describe doc maintenance as mostly manual. One harness-engineering skill defines drift as stale or missing edges between code and doc nodes and suggests matching doc text against current code entities. This is Delta's open niche.
7. **Skill evaluation exists as a field:** SkillsBench (87 tasks, deterministic verifier), AgentSkillOS (30 tasks), SkillAudit (per-artifact multi-dimensional assessment). Useful models for evaluating Delta's own skills.
8. **MSR 2026 study (466 OSS projects):** no established structure for AGENTS.md yet; projects continuously extend and maintain them.

### Sources

- https://www.augmentcode.com/tools/best-spec-driven-development-tools
- https://codex.danielvaughan.com/2026/05/22/spec-driven-development-frameworks-codex-cli-spec-kit-kiro-openspec-patterns-best-practices/
- https://felipefontoura.com/articles/spec-driven-development-tools-compared/
- https://reenbit.com/bmad-vs-spec-kit-vs-openspec-choosing-your-spec-driven-ai-framework/
- https://arxiviq.substack.com/p/evaluating-agentsmd-are-repository
- https://todatabeyond.substack.com/p/do-agentsmdclaudemd-files-help-coding
- https://arxiv.org/html/2510.21413v2
- https://mcp.directory/blog/claude-md-vs-agents-md-vs-skills-2026
- https://arxiv.org/pdf/2606.22613

---

## 4. Candidate directions

Legend — **Effort:** S/M/L. **Value:** how much it moves robustness, portability, or results.

### A. Measure before building (highest priority)

The ETH finding means Delta could be hurting agents today without anyone noticing. Everything else rests on knowing it helps.

- **A1. Delta evals.** A small benchmark of 10–20 real tasks. Compare three configurations: no KB, full KB, trimmed KB. Measure success, tokens, and step count. Decide from data. *Effort M, value very high.*
- ~~**A2. Context budget.** A line/token cap per doc and per spec, enforced by Archivist when writing. *Effort S, value high.*~~ **Done 2026-10-25 (Claude only).** Per-class ceilings in `.claude/budget.conf`, a read-only POSIX shell report (`budget.sh`, bundled with Archivist) and a rule. An over-budget artifact is fitted by lossless splitting through Archivist's step 8, never trimmed. First applied to Archivist, `ddd-core-layering` and the skill descriptions.
- ~~**A3. Load only what is needed.** Keep `CLAUDE.md` a minimal index; load docs and specs by path scope or on demand instead of always loading all six docs. *Effort S–M, value high.*~~ **Done 2026-10-25, reframed.** Checking how Claude Code loads context showed the six docs were never auto-loaded (only `CLAUDE.md`, rules without `paths:` and skill descriptions are), so the real problem is the agent not reading what it needs when it touches a slice. Delivered: the `CLAUDE.md` template is now an index of questions, with a line telling a session to list `deltas/` and open the matching spec before changing a capability, and it drops what the code already shows; the budget script's token estimate was calibrated to 3.3 characters per token. Set aside: a routing rule per slice (a file per slice to maintain that dies silently when code moves), a slice-to-code table (same upkeep), a nested `CLAUDE.md` importing the spec (imports load eagerly), and hooks. Revisit a per-slice rule only for a slice where sessions are seen skipping its spec.
- ~~**A4. Honest Surveyor.** Treat generated docs as hypotheses: keep only what is non-obvious and not derivable from the code, since derivable content is what the paper says distracts agents. *Effort S, value high.*~~ **Done 2026-10-25.** Archivist gained a "write only what the code can't tell" filter and a matching check before presenting a doc; Surveyor separates what it observed from what it inferred, asks for rationale and pitfalls, and states "not applicable here" in one line instead of padding a doc. The modules doc was refocused on cross-component flows, non-obvious behavior and shared guarantees, and the structure, database and expertise skeletons no longer invite copied lists or DDL. Not measured: whether this improves agent results (A1 stays deferred).

### B. Make Sentinel infrastructure (the differentiator)

- **B1. CI / pre-commit mode.** Run Sentinel on the slices touched by a diff; warn or fail on `Violated` or `Gone`. *Effort M, value very high.*
- ~~**B2. Verifiable invariants.** Each invariant carries a verification hint (test, command, grep). A deterministic part of Sentinel then needs no LLM, which is cheaper and more robust. *Effort M, value very high.*~~ **Done 2026-10-25, together with B3.** Every invariant and acceptance criterion in a spec ends with a verification tag: `[test: path "name"]`, `[check: command]`, `[manual: step]` or `[none]`. Specifier asks for it, archivist writes it and confirms a cited test exists, and sentinel runs the linked test or command instead of searching for one. The deterministic part is `links.sh`, beside sentinel: it counts claims by kind, lists the gaps and checks that every test link resolves, with no model.
- ~~**B3. Spec-to-test links.** Each acceptance criterion points to a test; no test means an explicit `Unverified`. *Effort S–M, value high.*~~ **Done 2026-10-25, together with B2.** Acceptance criteria carry the same tag, and no test is an explicit `Unverified` plus a gap in the report. Delta still never writes tests: a `[none]` tag makes the missing test visible under open questions, it does not supply it.
- **B4. Cheap reference check.** A no-LLM script that flags paths or symbols named in a spec that no longer exist. *Effort S, value high.* Partly covered: `links.sh` already checks the test files and names cited in verification tags, but not other paths or symbols a spec mentions.
- **B5. Scope Sentinel to docs/rules too (optional).** Today it audits only specs/designs; an extension could check stale references in docs. Beware of volatile references (docs deliberately avoid them). *Effort M, value medium.*

### C. Portability

- **C1. `AGENTS.md` as the canonical entry point**, with `CLAUDE.md` importing it. *Effort S, value high.*
- **C2. Single neutral source, generated variants.** Cursor support was removed on 2026-10-25, so no parallel copy is left to keep in sync. If another agent is targeted again, maintain one neutral format for skills and rules and generate each agent's variant with a small script. *Effort M, value medium (only if another agent is targeted).*
- **C3. Keep zero runtime.** This is the edge over Spec Kit. Any helper script must be a single file with no dependencies, and optional. *Constraint, not a task.*

### D. Bridge the Spec Kit gap without competing

- ~~**D1. Optional `implementer` skill.** Reads a spec, proposes a task plan, then triggers Sentinel. A light bridge, not a full `/implement`. *Effort M, value medium.*~~ **Done 2026-10-25, as `executor`.** An optional fifth skill that writes tests first and then code for one slice. Its work is computed from the claims whose verification tag is not satisfied yet (using `links.sh`), shown as a plan for approval and never stored. It edits source and tests only, hands tag updates and spec problems to archivist and specifier, and does not commit. Costs about 170 tokens of always-loaded description. Not yet run against a real agent.
- **D2. Change-delta mode.** A change proposes modifications to a spec (which invariant changes) before touching code. OpenSpec's good idea, but the contract remains the only artifact. *Effort M, value medium–high.*
- **D3. Interoperability.** Import/export between Spec Kit or OpenSpec specs and Delta's format. *Effort M, value medium.*

### E. Spec quality

- ~~**E1. Stricter invariant notation** (EARS or Given/When/Then) so Sentinel judgments are less ambiguous. *Effort S–M, value high.*~~ **Done 2026-10-25.** Invariants are one claim per line in an EARS form (always/never, When, While, If-then, Where), in plain language with an observable response; acceptance criteria are Given/When/Then lines; design States and Interactions follow the same forms. Required for new or revised specs, older ones stay until touched. Specifier rewrites vague claims, archivist checks the form, sentinel reads the trigger and response to know what to exercise, executor maps a claim onto a test. No lint, since a keyword check would be fragile. Not measured.
- **E2. Persistent open questions** in Specifier that block closing a spec until resolved. *Effort S, value medium.*
- **E3. Per-slice health metrics:** last verified date, percent of invariants verified, open questions. *Effort M, value medium.*

### F. Speculative

- **F1. Incremental Surveyor:** rescan and propose only the KB diff instead of regenerating. *Effort M.*
- **F2. Orphan/overlap detection** for slices, automating the manual criteria in `when-write-delta`. *Effort M.*
- **F3. Hooks:** after a session that touched a slice, suggest running Sentinel. *Effort S.*

## 5. Suggested order

1. **A** first, even a small version. If Delta does not beat "no KB" on our own tasks, the rest is built on sand.
2. **B** next; it is the differentiator and the deterministic part makes it robust and cheap.
3. **C** in parallel; cheap and opens the audience.
4. **D** only if A and B come out well.
5. **E** and **F** as appetite allows.

## 6. Open decisions

- **Naming collision:** OpenSpec's "delta" format overlaps with the Delta name and may confuse search and positioning.
- **Audience:** personal/team harness or a publicly released project? This changes the priority of C (portability) and of A1 (public evals).
- **Eval design:** which real tasks, which repos, which agent/model, and what counts as success.
- **Write funnel:** is "everything through Archivist" worth its friction once CI-driven checks exist?
- **Scope of Sentinel:** keep it to specs/designs, or extend to docs and rules?
