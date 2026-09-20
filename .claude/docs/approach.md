<!--
This file ships as part of the harness itself, as the fill-in skeleton for
this doc category — copy it into a project as-is, then fill it in place
when bootstrapping (surveyor) or updating (archivist) that project's build
approach. Unlike structure.md, this file is always created for a
bootstrapped project, even when the honest content is "not applicable here"
(no meaningful build sequencing beyond "just build the thing") — state that
explicitly rather than omitting the file. This file is never discarded once
work lands — it stays live-edited the same way every other doc does: update
the roadmap and status in place as phases complete, don't leave a finished
phase's entry describing it as pending.
Delete these guidance comments once every section holds real, grounded
content. Ground every claim in what was actually decided/confirmed, never
invent a phase or pillar that hasn't actually been agreed on.
-->

# Approach

<!-- One sentence: what this file covers — the order and philosophy behind building this project — and what it deliberately excludes (e.g. "Why the project exists lives in overview.md, not here"). -->

---

## Technical pillars

<!-- Table: the few things this project deliberately builds together because the problem genuinely needs them at once, not because they were bolted on separately. Explain *why together*, not just list them. Omit if the project has no such deliberate combination — most don't. -->

| Pillar | What it means here |
|---|---|
| `<pillar>` | `<why this matters and why it's tied to the others>` |

## Functional scope

<!-- What the system being built must actually support, as a bullet list of concrete capabilities — not a feature-marketing list, a scope boundary a build sequence can be checked against. -->

## Roadmap

<!-- Ordered phases/stages the build proceeds through, each with what it actually establishes. Update this in place as phases complete — mark a phase done when it's done, don't leave it worded as future work after it lands. -->

0. **`<phase name>`** — `<what this phase establishes and why it comes first/next>`

## Risks

<!-- Known risks to the roadmap actually landing as planned — an unresolved external dependency, an unvalidated scaling/performance assumption, a technology bet that might not pan out. Not a generic risk register: only a risk specific and real to this roadmap. Omit entirely if there genuinely isn't one. -->

## Done criteria

<!-- How to tell the current scope (or a named phase of it) is actually done — specific, falsifiable conditions, not a vibe. If it can't be checked by reading the resulting code/behavior, it isn't specific enough yet. -->

## Stretch goals

<!-- Explicitly non-blocking extensions — real candidates, not blocking scope, deferred once the done criteria above are actually met. Omit if there are none. -->

---

## Non-goals

<!-- Only include this section if a scope boundary here is easy to violate by accident. Omit entirely otherwise. -->
