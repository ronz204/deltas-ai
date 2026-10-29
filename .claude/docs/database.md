<!--
Fill-in skeleton shipped with the harness: copy it into a project and fill it in
place (surveyor to bootstrap, archivist to update). Delete these comments once
every section holds real, grounded content. Ground every claim in the real
schema, migrations and provisioning config, never in what a similar stack
usually looks like.
Record what the schema files don't say on their own: rationale, invariants,
access split. Don't copy DDL the migrations already hold. The first two sections
are the ones worth keeping; the rest are optional and only for what the schema
doesn't make plain. Omit any section that doesn't apply to this store.
-->

# Database

<!-- One sentence: what this file covers (the rules the data must obey and who may touch it) and what it excludes (e.g. "Application-level topology lives in structure.md"). -->

---

## Persistence invariants

<!-- What must hold about the data regardless of which code path writes it: uniqueness, referential integrity, a cross-entity consistency guarantee, a rule a migration must never violate. Name and explain any non-obvious enforcement pattern relied on (e.g. a composite key trick), not just the invariant. This is what gets checked against a future schema change. -->

## Access control

<!-- Who or what can touch the store and how narrowly: roles or credentials split by purpose (migration vs. runtime, read vs. write) and any row-level or per-tenant isolation the store enforces itself. A table (Role | Purpose | Used by) fits the role split. State *why* the split exists, what it prevents. -->

## Schema

<!-- Optional. How objects are namespaced and why: a dedicated schema vs. the engine default, a naming convention for collections. -->

## Data model

<!-- Optional. Entities and relationships in domain terms, plus constrained value types, only where the schema files don't make that plain. A fenced block is for a structural shape the schema doesn't state. Describe the shape, never the source-tree location that defines it. -->

## Infrastructure

<!-- Optional. Engine, hosting model, backup and replication: static facts set up once. Named roles are fine; a connection string, credential or file path is not. -->

## Access patterns

<!-- Optional. The convention for reads, writes and migrations, and any pattern meant to avoid a known failure mode (unbounded fetches, N+1). Explain why it exists. -->

---

## Non-goals

<!-- Only if a scope boundary here is easy to violate by accident. Omit otherwise. -->
