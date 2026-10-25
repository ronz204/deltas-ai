<!--
This file ships as part of the harness itself, as the fill-in skeleton for
this doc category — copy it into a project as-is, then fill it in place
when bootstrapping (surveyor) or updating (archivist) that project's
persistence reference. Unlike structure.md, this file is always
created for a bootstrapped project, even when the honest content is "not
applicable here" (no persistent store) — state that explicitly rather than
omitting the file, same discipline surveyor's own baseline checklist uses
for security/performance/scalability. Data model and database
infrastructure are kept together here, deliberately, rather than split
against structure.md. Delete these guidance comments
once every section holds real, grounded content. Ground every claim in the
real schema/migrations/provisioning config (Step 0), never from memory or
from what a similar stack usually looks like. Omit any section below that
genuinely doesn't apply to this project's store (e.g. Access control's
row-level policies on a store with no such feature) rather than forcing
content into it.
Record what the schema files don't say on their own: the rationale, the
invariants, the infrastructure. Don't copy DDL that the migrations already hold.
-->

# Database

<!-- One sentence: what this file covers. If genuinely not applicable (no persistent store), say so plainly here and skip the remaining sections rather than forcing content into them. -->

---

## Schema

<!-- How objects are namespaced/organized at the store level, and why — a dedicated schema/namespace vs. the engine's default, a naming convention for collections. Omit if the store has no such concept. -->

## Access control

<!-- Who/what can touch this store and how narrowly — roles or credentials split by purpose (migration vs. runtime, read vs. write), and any row-level or per-tenant isolation policy enforced by the store itself rather than trusted to application code. A table (Role | Purpose | Used by) is usually the clearest shape for the role split. State *why* the split exists (what it prevents), not just that it does. -->

## Data model

<!-- The entities that matter and how they relate in domain terms, plus the value types (enums, domain-specific types) that constrain them — only where the schema files don't make that plain on their own; don't copy DDL the migrations already hold. A fenced code block is for a structural shape the schema doesn't state outright. Entity/field names are durable architecture vocabulary — describe the shape, never the source-tree location that defines it. -->

## Persistence invariants

<!-- What must always hold about the data regardless of which code path writes it — uniqueness, referential integrity, a cross-entity consistency guarantee, an invariant a migration must never violate. Name and explain any non-obvious enforcement pattern relied on to hold one of these (e.g. a composite key trick), not just the invariant itself. This is usually the highest-value section: it's what actually gets checked against a future schema change. -->

## Infrastructure

<!-- Engine, hosting model, backup/replication approach — static facts set up once and rarely revisited. Named infrastructure roles are durable architecture vocabulary and safe to name; a specific connection string, credential, or file path is not. -->

## Access patterns

<!-- How the application actually talks to this store — the convention for reads/writes, migrations, and any pattern meant to avoid a known failure mode (e.g. unbounded fetches, N+1 queries). Explain why the convention exists, not just what it is. -->

---

## Non-goals

<!-- Only include this section if a scope boundary here is easy to violate by accident. Omit entirely otherwise. -->
