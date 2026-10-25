<!--
This file ships as part of the harness itself, as the fill-in skeleton for
this doc category — copy it into a project as-is, then fill it in place
when bootstrapping (surveyor) or updating (archivist) that project's
mechanism reference. Unlike structure.md, this file is always created
for a bootstrapped project, even when the honest content is "not applicable
here" (nothing here relies on a mechanism worth explaining) — state that
explicitly rather than omitting the file. Delete these guidance comments
once every section holds real, grounded content. Ground every claim in the
real technology/algorithm/protocol actually in play (Step 0) or in
established domain fact — never invent a mechanism explanation that hasn't
actually been confirmed.
Explain only the mechanisms where a misunderstanding would cause a mistake in
this project; general knowledge a session already has doesn't belong here.
-->

# Expertise

<!-- One sentence: this file explains *how* a non-trivial mechanism or concept the project depends on actually works — not *whether*, *when*, or *where* it's used; that's structure.md's or database.md's job. If genuinely not applicable, say so plainly here and skip the remaining sections. -->

---

## Glossary

<!-- Optional. Only for terse, domain- or technology-specific terms that need a one-line definition before the sections below make sense — not a substitute for the sections themselves. Table: term -> what it means here. Omit entirely if nothing needs this. -->

| Term | Meaning |
|---|---|
| `<term>` | `<what it means here>` |

## `<Mechanism or concept name>`

<!--
Repeat this section per mechanism/concept that this project's design leans
on and that a reader can't be assumed to already know cold — a database
feature, a protocol behavior, a non-obvious algorithm/data-structure
property, a domain rule with real consequences if misunderstood. Explain
the underlying behavior in enough depth that a reader could reason about it
correctly without re-deriving it from source: what it actually does, what
guarantee it does/doesn't provide, and the caveat that trips people up.
Prose is usually right here; use a table only when comparing genuine
variants of the same mechanism (e.g. competing strategies for the same
problem).
-->

## `<Next mechanism or concept name>`

---

## Non-goals

<!-- Only include this section if a scope boundary here is easy to violate by accident. Omit entirely otherwise. -->
