<!--
This file ships as part of the harness itself, as the fill-in skeleton for
this doc category — copy it into a project as-is, then fill it in place
when bootstrapping (surveyor) or updating (archivist) that project's
cross-component behavior reference. It holds only what reading one component
doesn't reveal: a component's own step-by-step behavior and data shape are in
its code, and a capability's contract is its spec. Delete these guidance
comments once every section holds real, grounded content. Ground every claim in
the real source (Step 0) — never infer behavior from a filename or from what a
similar project usually does. Omit a section with nothing real to say.
-->

# Modules

<!-- One sentence: what this file covers — behavior that spans components or would surprise a reader — and what it deliberately excludes (e.g. "A component's own behavior lives in its code and a capability's contract in its spec, not here"). -->

---

## Cross-component flows

<!-- A flow that crosses more than one component, where the order or the hand-off isn't visible from any single one of them: a numbered list or a short sequence naming the components by role. Describe what really happens, not an idealized version. Component names are durable architecture vocabulary — describe the boundary, not the source-tree location that implements it. -->

## Non-obvious behavior

<!-- Per component, only behavior that would surprise a reader or break if changed: hidden coupling, an ordering requirement, a side effect, state kept implicitly. For each, name the component, what it does, and why it matters. Skip anything a reader gets by reading that component. -->

## Shared guarantees

<!-- What several components rely on together that no single one states — an ordering, an idempotency expectation, a shared assumption. A guarantee that belongs to one capability goes in that capability's spec instead. Use a fenced code block for anything structural. -->

---

## Non-goals

<!-- Only include this section if a scope boundary here is easy to violate by accident. Omit entirely otherwise. -->
