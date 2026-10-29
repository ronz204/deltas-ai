<!--
Fill-in skeleton shipped with the harness: copy it into a project and fill it in
place (surveyor to bootstrap, archivist to update). Delete these comments once
every section holds real, grounded content. Ground every claim in the actual
stack and config, never in what a similar project usually looks like.
Record the reasoning and the facts the code doesn't show. A technology list, a
dependency map or a request flow that manifests and imports already state is
pointed at in general terms, never copied. A flow whose order isn't visible from
one component goes in modules.md.
-->

# Structure

<!-- One sentence: what this file covers (the reasoning behind the architecture) and what it excludes (e.g. "The data model lives in database.md, not here"). -->

---

## Decisions and rationale

<!-- Table: only the choices whose reason the manifests and config don't show: a stack pick, a hosting model, a boundary, a rejected alternative. Skip any row a reader gets by opening the manifest. -->

| Decision | Choice | Why |
|---|---|---|
| `<decision>` | `<what was chosen>` | `<the reason, and what it rules out>` |

## Infrastructure

<!-- Static facts outside the persistence layer: hosting, environments, provisioned resources (queues, caches, CDNs, compute) and why each was provisioned that way. The database's own engine, hosting and backup live in database.md. Omit if the project runs as a single local process. -->

## Cross-cutting patterns

<!-- Patterns that apply across components: auth, background jobs, caching, rate limiting, error handling. State each with its rationale, not just its mechanics. -->

## Open architecture decisions

<!-- A choice deliberately not made yet and what resolving it depends on: a real trade-off being carried, not a task waiting to be scheduled. Once decided, fold it into the sections above and remove it from here. Omit when nothing is open. -->

---

## Non-goals

<!-- Only if a scope boundary here is easy to violate by accident. Omit otherwise. -->
