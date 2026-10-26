<!--
Surveyor's own working checklist for Step 2 — not something that gets copied
into a project's knowledge base, and not living truth that needs updating
per project. Every item below is a question, never an answer: confirm or
investigate each against this project's real stack before recording
anything, same Step 0 discipline as every other skill in this family.
-->

# Baseline Checklist — Security, Performance, Scalability

Three concerns every project has to satisfy in some form, even when the answer is "not applicable here." Transversal as questions, never as prewritten answers — the concrete answer always depends on this project's real stack, so nothing below gets written into a project's docs/rules as-is.

---

## How to use this

- Walk through the sections below during Step 2, alongside the vision/scope/rule-candidate questions — same short, focused-batch discipline.
- Route what gets confirmed by what it actually is, not by which section it came from:
  - A durable architectural decision (an auth chokepoint, a caching strategy, a stateless-by-design constraint) → a note for `.claude/docs/structure.md`'s cross-cutting-patterns content.
  - An always-apply convention (validate all input at the boundary, never log secrets, paginate every list endpoint) → a rule candidate, confirmed the same way any other rule candidate is (see `archivist`'s "Spotting a rule candidate").
  - A guarantee scoped to one specific slice (this endpoint never processes a payment without a valid session) isn't surveyor's to write — flag it in Step 4's report as a consideration for `specifier` to fold into that slice's `spec.md` Invariants when it actually gets specced.
- An item that genuinely doesn't apply (no persistent data, no public network surface, single-user local tool) gets marked not-applicable and dropped — don't force an answer to fill the section.

## Security

- Where does authentication happen, and is there one chokepoint or several scattered checks?
- Where does untrusted input get validated or sanitized, and against which class of injection (SQL, command, template, XSS)?
- How are secrets and credentials stored and accessed — is there any path where one could end up hardcoded or logged?
- Does the project handle sensitive or regulated data, and if so, what handling constraint follows (encryption at rest/in transit, logging exclusions, retention limits)?
- Is there a policy, even an informal one, for vetting a third-party dependency before adding it?

## Performance

- Which operations are latency-sensitive, and is there a stated target or budget for them?
- Does the data-access pattern risk unbounded fetches or N+1 queries, and is there a project convention meant to avoid it?
- Is anything cached, and if so, what invalidates it?
- Do list-returning endpoints or queries paginate by default, or is that decided case by case?

## Scalability

- Can a single running instance be killed and replaced without losing state, or is there in-memory state that wouldn't survive a restart?
- How is long-running or deferred work handled — a queue/worker, or inline in the request path?
- Is there a data shape expected to grow large enough that partitioning, sharding, or archiving will eventually matter?
- Is there a rate-limiting or throttling policy for any publicly reachable surface?

---

## Non-goals

- Not a checklist of guaranteed problems — an item that doesn't apply to this project is dropped, not forced into an answer.
- Not enforced directly by `sentinel` — only what a confirmed answer becomes (a rule, a `structure.md` entry, or a later slice's spec invariant) is ever checked again; this file itself is never re-read as a project's living truth.
