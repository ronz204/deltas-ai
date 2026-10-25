---
paths:
  - "**/source/core/**"
---

# Bounded Context Layering Conventions

This governs the internal folder and file layout of every bounded context under `source/core/` (how each building block is written is a separate rule), following a DDD tactical layering organized by semantic sub-slice: each bounded context is split into the cohesive concepts it holds, not into technical layers. It applies to every bounded context there, including the one holding primitives shared across the others. Where it needs a language's own mechanism (file extension, visibility, how a type is made immutable or distinct, how an interface is marked), the language idioms rule for the file's extension supplies it. The files inside a sub-slice (their fixed kinds, repositories, and role contracts) are the sub-slice anatomy rule's concern.

---

## Bounded context folders

- Each bounded context is its own folder directly under `source/core/`, named as a two-word pair (joined the way the language's package names allow), because a shared naming shape makes the set of bounded contexts scannable as a flat list rather than a mix of single words and long compounds.
- The second word is chosen to be descriptive of that specific context, not a suffix repeated across every bounded context, because a repeated suffix (e.g. always `-domain` or always `-core`) stops carrying information the moment every folder has it — it becomes noise instead of a signal.

## What core may depend on

- A bounded context imports only the shared-kernel bounded context, never another bounded context. Collaboration between contexts happens in the application layer, which may use several of them. Letting contexts import each other in core is how they end up in import cycles and coupled through their internals, so that neither can be understood or tested on its own.
- Core never imports the application layer, infrastructure, persistence, runtime configuration, or any framework or I/O library. The dependency direction is what keeps the domain replaceable around its adapters: once core imports one adapter-side module, every change on that side can break the domain. When core needs something from that side, it declares a contract instead.

## Semantic sub-slices

- A bounded context splits into one subfolder per sub-slice — a cohesive domain concept, typically one entity (the root of an aggregate) plus everything that exists only to serve it (its parts, value objects, policies, events, errors, repository, and role contracts). The sub-slice folder is named with a single short word naming that concept from inside the bounded context, because the bounded context folder already supplies the qualifier: a lifecycle context holding a definition and its executions gets `definition/` and `execution/`, not a repeat of the context's own noun in front of each.
- There is no split by technical layer (no separate folder for the domain model versus the interfaces it needs from outside). Grouping by layer scatters one concept across two trees and forces every change to that concept to touch both; grouping by concept keeps everything that changes together in one folder.
- A shared-kernel-style bounded context (primitives shared across bounded contexts: base value objects, shared ids, the base domain error, shared contracts) follows the exact same shape — its primitives are grouped into sub-slices by concept (e.g. everything about identifiers in one sub-slice), with no separate carve-out. A format value object several contexts share (an email address) sits with the primitive factories instead of getting a sub-slice of its own, because a sub-slice holding one value object and its error is a folder with nothing to group.
- Anything every bounded context needs lives in the shared kernel, not in the context that happens to own the data behind it. Admission to a tenant (membership standing, required second factor, permission check) is the example: every context's endpoints run it, so its policy and the read model that policy decides on are shared-kernel, and the context that owns the membership tables implements the reads and the use case that runs the admission.
- Sub-slices are always a single level deep — a sub-slice never nests further subfolders. If a concept grows enough to want internal structure, that's a signal it's really two sub-slices side by side, not one with a nested tree.
- Every bounded context uses sub-slice folders, even one holding a single concept today, because deciding folder shape per context based on how many concepts it currently has is exactly the kind of ad hoc call that drifts silently as the context grows.
- A sub-slice is as large as its concept needs. A big entity with many policies and value objects is essential complexity of a rich domain and stays one sub-slice; splitting it only to shrink the folder separates rules that change together. The split signal is two concepts, never size.

```
source/core/
  <bounded-context>/
    <public entry point>
    <subslice-a>/
      <subslice-a>.entity
      <subslice-a>.parts
      <subslice-a>.vos
      <subslice-a>.policies
      <subslice-a>.events
      <subslice-a>.errors
      <subslice-a>.repository
      <subslice-a>.notifier
      <subslice-a>.scheduler
      ...
    <subslice-b>/
      ...
```

## The barrel is the only public surface

- Each bounded context has a public entry point at its own root (its barrel), acting as its single way in: it re-exports the public symbols of every file of every sub-slice. How the entry point is written comes from the language idioms rule. Nothing outside the bounded context imports from a sub-slice folder directly, because that would let other contexts (or the application layer) couple to internal file tenant that is free to change without notice.
- The language's own visibility mechanism is the public/private gate: a symbol that must stay internal to the bounded context isn't made public in its own file, rather than being filtered out by hand at the barrel. Whether a new public symbol also needs an edit at the barrel depends on the language, and the language idioms rule says which.
- Because every sub-slice's exports land in one flat namespace at the barrel, exported symbol names must be unique across the whole bounded context. Names carry their full domain noun (e.g. a definition sub-slice's id type is named after the full concept, not a bare `Id`) rather than relying on the folder for disambiguation, since the folder is invisible to importers.

## Building blocks

- How each tactical piece is written (value objects, identifiers, literal sets, errors, entities, parts, events) is its own convention, in the domain building-blocks rule, so this file stays about where things live and that one about how they are shaped.

---

## Non-goals

- This rule governs internal layering only: where things live and what may import what. It does not decide which domain concepts belong to which bounded context, how a bounded context is carved into sub-slices, or where one entity's aggregate boundary ends and another's begins — those are per-bounded-context decisions made in that slice's own `spec.md`, not here.
- It doesn't govern the application or infrastructure layers' own layout — those are separate conventions with their own trigger paths.
