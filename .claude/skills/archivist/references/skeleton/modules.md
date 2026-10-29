<!--
Skeleton for a .claude/docs doc. Create the doc from it only when something
passes the filter (what the code and README don't show); a category with nothing
to say is not created. Delete these comments once every section holds real,
grounded content. Ground every claim in the real
source, never infer behavior from a filename or from what a similar project does.
It holds only what reading one component doesn't reveal: a component's own
behavior and data shape are in its code, and a capability's contract is its
spec. Omit a section with nothing real to say.
-->

# Modules

<!-- One sentence: what this file covers (behavior that spans components or would surprise a reader) and what it excludes (e.g. "A component's own behavior lives in its code and a capability's contract in its spec, not here"). -->

---

## Cross-component flows

<!-- A flow that crosses more than one component where the order or the hand-off isn't visible from any single one: a numbered list or short sequence naming components by role. Describe what really happens, not an idealized version. Name the boundary, not the source-tree location. -->

## Non-obvious behavior

<!-- Per component, only behavior that would surprise a reader or break if changed: hidden coupling, an ordering requirement, a side effect, state kept implicitly. Name the component, what it does and why it matters. -->

## Shared guarantees

<!-- What several components rely on together that no single one states: an ordering, an idempotency expectation, a shared assumption. A guarantee that belongs to one capability goes in that capability's spec. Use a fenced block for anything structural. -->

---

## Non-goals

<!-- Only if a scope boundary here is easy to violate by accident. Omit otherwise. -->
