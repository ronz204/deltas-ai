---
paths:
  - "*/core/**"
---

# Bounded Context Layering Conventions

This governs the internal folder and file layout of every bounded context under `source/core/`, following a hexagonal-plus-DDD tactical layering. A shared-kernel folder under `source/core/` already holds real primitives; bounded-context aggregates themselves are not implemented yet beyond empty scaffolding, so this convention is forward-looking for that part, corrected against the folder/file naming the shared kernel already uses in practice.

---

## Bounded context folders

- Each bounded context is its own folder directly under `source/core/`, named as a two-word kebab-case pair, because a shared naming shape makes the set of bounded contexts scannable as a flat list rather than a mix of single words and long compounds.
- The second word is chosen to be descriptive of that specific context, not a suffix repeated across every bounded context, because a repeated suffix (e.g. always `-domain` or always `-core`) stops carrying information the moment every folder has it — it becomes noise instead of a signal.

## Contexts and contracts split

- Every bounded context folder splits into exactly two top-level subfolders: `contexts/`, holding the domain model itself (aggregates, entities, value objects, domain events, domain errors), and `contracts/`, holding the ports that bounded context exposes outward (repository interfaces, DTOs, published-event contracts) for other contexts or the application layer to depend on.
- Nothing outside a bounded context imports from its `contexts/` or `contracts/` folders directly — `index.ts` is the only supported dependency surface, because letting outside code reach into either folder directly would let another bounded context (or the app layer) couple to internal representations that are free to change without notice.
- `contracts/` never holds a re-export of a concrete aggregate, entity, or domain-error class merely to make it reachable from outside — only genuine port/interface/DTO declarations belong there (repository interfaces, published-event contracts, request/response DTOs). A bounded context with no ports yet has an empty, or absent, `contracts/` folder — that's not a gap to fill preemptively with re-exports of whatever `contexts/` happens to contain.
- Each bounded context also has an `index.ts` at its own root, acting as its single public entry point and the one place outside `contexts/` and `contracts/` allowed to import from either: it re-exports every file in `contexts/` (and `contracts/`, where that folder exists) via `export * from "./contexts/<file>"` (or `./contracts/<file>`) — one line per file that exports at least one symbol. Other code imports the bounded context through `index.ts` — never by reaching into `contracts/`'s own files, or into `contexts/`, directly.
- The `export` keyword inside a concrete file under `contexts/` or `contracts/` is the public/private gate, not the barrel — a symbol that must stay internal to the bounded context simply isn't given an `export` in its own file, rather than being filtered out by hand in `index.ts`. This trades barrel-level curation for trusting that gate: once a file exports something, it's meant to be reachable from outside and flows through `export *` automatically, so `index.ts` never needs to change just because a file gains a new export — that decision was already made at the point where `export` was written.

## contexts/ is flat

- `contexts/` never nests a per-aggregate subfolder — every file, across every slice a bounded context holds, sits flat directly under `contexts/`. A bounded context routinely holds more than one slice (aggregate) side by side; flat layout doesn't imply one aggregate per bounded context, it just means slices share the folder instead of each claiming its own.
- This applies uniformly to every bounded context, including a folder under `source/core/` whose purpose is to hold primitives shared across bounded contexts (base value objects, shared ids, common base types, shared domain errors) rather than an aggregate of its own — that folder keeps the same `contexts/` folder and `index.ts` barrel as any other bounded context, on the same flat terms, with no separate carve-out needed.
- Flat-vs-nested is deliberately not a per-bounded-context decision based on how many slices it happens to have today — that's exactly the kind of ad hoc call that drifts silently as a context gains slices over time. A single uniform rule removes the decision entirely, and the file-naming convention below is what keeps multiple slices' files unambiguous side by side in the same folder.
- A shared-kernel-style folder typically also has no `contracts/` folder — it's shared by direct reference in DDD, not translated across a port like a real bounded context's internals — and its `index.ts` barrel re-exports `contexts/` directly rather than a `contracts/` translation layer. The `contracts/` folder remains available on the same terms as any bounded context if a shared-kernel concept ever does need one.

## File naming inside contexts/

- Inside `contexts/`, files are named `<slice>.aggregate.ts`, `<slice>.entities.ts`, `<slice>.vos.ts`, `<slice>.events.ts`, `<slice>.errors.ts`, `<slice>.enums.ts` (a closed set of literal values, e.g. a state or tier enum), and `<slice>.types.ts` (other type aliases/interfaces the slice needs that don't fit any of the other file kinds, e.g. the shape of a persisted snapshot a reconstitution factory accepts) — only the ones that slice actually needs (a slice with no domain events yet has no `.events.ts` file), because an empty placeholder file communicates nothing a missing file doesn't already communicate.
- The `<slice>.` prefix is what disambiguates which aggregate a file belongs to now that there's no subfolder doing that job — it's load-bearing, not cosmetic, since a flat `contexts/` folder relies on it to keep two slices' `*.errors.ts` (or any other same-kind file) from colliding.

## Aggregate immutability

- An aggregate favors `readonly` public fields over a private field paired with a getter, when the getter would do nothing but return that field — that getter is boilerplate, not encapsulation, since it adds a method without adding any actual guard or transformation.
- A state-changing operation on an aggregate returns a new instance rather than mutating the receiver in place, so a caller still holding a reference to the pre-change instance keeps a value that stays valid and unaffected by the change, instead of having it silently mutate out from under them.

---

## Non-goals

- This rule governs internal layering and file naming only. It does not decide which domain concepts belong to which bounded context, or where one aggregate's boundary ends and another's begins — those are per-bounded-context decisions made in that slice's own `spec.md`, not here.
