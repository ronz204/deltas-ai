<!--
This file ships as part of the harness itself, as the fill-in skeleton for
this doc category — copy it into a project as-is, then fill it in place
when bootstrapping (surveyor) or updating (archivist) that project's
topology reference. Delete these guidance comments once every section holds
real, grounded content. Ground every claim in the actual stack/config
(Step 0), never in what a similar project usually looks like.
Record the reasoning and the facts the code doesn't show. A technology list or a
dependency map that the manifests and imports already state is pointed at in
general terms, never copied here.
-->

# Structure

<!-- One sentence: what this file covers and what it deliberately excludes (e.g. "The data model itself lives in database.md, not here"). -->

---

## Stack

<!-- Table: each major technology choice and why it was made — not just what was picked. -->

| Layer | Choice | Rationale |
|---|---|---|
| `<layer>` | `<technology>` | `<why this one>` |

## Topology

<!-- How the parts actually communicate — request flow, sync vs. async boundaries, what talks to what. Prefer a short structural breakdown or an ASCII diagram in a fenced code block over prose when the shape is spatial. -->

## Infrastructure

<!-- Static infrastructure/provisioning facts outside the persistence layer — hosting model, environments, provisioned resources (queues, caches, CDNs, compute) — and why each was provisioned that way. Persistence-layer infrastructure (the database's own engine/hosting/backup approach) lives in database.md instead, not here. Omit if the project genuinely has none of this (e.g. runs as a single local process). -->

## Cross-cutting patterns

<!-- Patterns that apply across components rather than to one of them specifically — auth, background jobs, caching, rate limiting, error handling. State each pattern with its rationale, not just its mechanics. -->

## Open architecture decisions

<!-- A choice deliberately not made yet, and what resolving it depends on — a real trade-off being carried, not a task waiting to be scheduled. Once decided, fold the choice and its rationale into Stack/Topology/Cross-cutting patterns above and remove it from here; this section is never a place a resolved decision keeps living. Omit entirely when there's nothing genuinely open. -->

---

## Non-goals

<!-- Only include this section if a scope boundary here is easy to violate by accident. Omit entirely otherwise. -->
