---
paths:
  - "**/source/core/**"
  - "**/source/repos/**"
---

# Domain Building Blocks Conventions

This governs how each tactical DDD piece in the domain layer is written: value objects, identifiers, literal sets, domain errors, entities, aggregates and domain events, plus the mapper that rebuilds an aggregate from a row. Every piece is built from one base factory or type in the shared kernel, so two aggregates or two value objects never differ in shape, only in content. Where these pieces live (folders, file kinds, the barrel) is the layering rule's concern, not this one's.

---

## One formula per piece

| Piece | Built from | Shape it always has |
|---|---|---|
| Value object | The shared kernel's brand type plus its text factory (or a hand-written parser for a non-text value) | A same-named constant with `kind`, `parse` and `is`, and a branded type inferred from it |
| Identifier | The shared kernel's identifier factory, one declaration per id kind | Same as a value object; ids are lowercase UUIDs |
| Literal set | The shared kernel's literal-set factory | A constant with `kind`, `values`, `parse` and `is`, and a type derived from its values |
| Domain error | The base domain error's `define` factory, one class per sub-slice | A closed set of `snake_case` codes with fixed messages |
| Entity | The shared kernel's entity base, typed by a state interface that extends the base entity state | Frozen instance, `restore`, `sameAs`, protected `with` |
| Aggregate | The same entity base: a root is an entity, and there is no separate aggregate base | Same as an entity, plus intention-named creation factories |
| Domain event | The shared kernel's domain event type | `type`, `aggregateId`, `occurredAt`, `payload` |

A piece that seems to need a shape outside this table is a signal to extend the shared-kernel base once, for every context, not to hand-roll a variant in one sub-slice. A hand-rolled variant is how two aggregates end up with two different ways to rebuild from storage.

## Value objects, identifiers and literal sets

- A value object is a branded primitive with a same-named constant holding its parser. The parser is its only constructor: it normalizes, validates and raises the sub-slice's domain error. The brand keeps two string-backed concepts from being swapped silently, while equality stays `===` and the value maps to a column with no wrapping. The type is inferred from the constant rather than written by hand, so the kind is declared once and the type cannot drift from its parser.
- A text value object is declared through the text factory with its bounds, an optional pattern and an optional normalization, and the error to raise, which must be a domain error and receives the bounds as structured details. Trimming always happens first, so no value object re-implements it.
- `is` is a pure guard: it accepts only a value already in normalized form, so a guard never reports `true` for a value `parse` would have changed.
- A closed set of literals (a role, a status) goes through the literal-set factory, and its type is derived from the set's values, so the type and the runtime list can't drift. A set whose rejection is part of the API contract passes its own error to the factory; otherwise the shared kernel's invalid-literal error is raised.
- A class-based value object with a wrapped field and an `equals` method is not used: it adds an unwrap at every boundary and an equality method the brand already gives for free.

## Domain errors

- Every domain error extends the shared kernel's base domain error and carries a stable `snake_case` code. Codes are part of the API's stable contract, so the HTTP layer maps a code to a status without knowing each error class, and renaming a class or rewording a message never breaks a client.
- Each sub-slice has one error class, declared by extending what the base error's `define` factory returns for a map of code to message. Its code type is derived from that map, so a status map keyed by it fails to compile when a code has no status.
- The message is fixed per code, in English, and never interpolates a value. Variable data (a limit, a current state) travels in the error's structured details, which keeps monitoring grouping one failure as one issue and keeps user-supplied values out of logs.

## Entities and aggregates

- The state is a named interface extending the base entity state, so every entity and aggregate has an `id` the base can compare. Optional state traits from the shared kernel (a version for optimistic concurrency, creation and update timestamps, a soft-delete timestamp) are mixed into that interface only when the table really has those columns, because a trait without a column is a field the mapper has to invent.
- The base assigns the state and freezes the instance, so fields are `readonly` in the type system and at runtime, and the class never repeats its field list in a constructor, a getter or a serializer.
- The constructor is private and takes the whole state object, never positional arguments, because two positional arguments of the same type swap silently.
- Creation goes through intention-named static factories that all take `(input, now)`: a named input object plus the clock. Core never generates ids or reads the time, so the caller passes both and a factory stays deterministic.
- Reconstitution goes through the `restore` factory every entity and aggregate inherits from the base. It trusts its input, because the mapper already parsed every field; no aggregate writes its own.
- Identity comparison goes through the inherited `sameAs`, which compares class and id, never through `===` on instances, because two restores of the same row are different objects that are the same entity.
- No getters and no snapshot method. A field is read directly; a derived fact is a method whose name says it computes (`isActive()`).
- A state-changing operation returns a new instance through the base's protected `with` patch, never a mutation of the receiver and never a hand-written copy. A patch never carries the id, so no operation can change an entity's identity. An operation that leaves the state unchanged may return the receiver itself, so a caller can skip the write by identity.

## Domain events

- An event is a value of the shared kernel's domain event type, with a literal `type` named `<aggregate>.<past_tense_verb>`, the aggregate's id, the time passed in by the caller, and a payload of plain values.
- An operation that raises events returns the shared kernel's change result, the new instance as `next` plus its `events`, and never publishes anything itself. The aggregate can't know whether its change will commit, so only the caller, which owns the transaction, can hand the events to the outbox in the same run.
- An operation that raises no event returns the bare instance, not an empty change result, so the signature says whether events exist.

## Mappers

- The mapper beside a repository has exactly two directions: to the row for writes, and to the domain for reads. To the domain, it parses every column with the same value-object parsers the domain uses and then calls `restore`, so a row that drifted from the domain's rules fails on read instead of flowing in.
- The mapper is the only place that knows both the row type and the state interface. A repository never builds an aggregate or a row by hand.

---

## Non-goals

- Doesn't decide which aggregates a context has or where their boundaries are; that belongs to each slice's spec.
- Doesn't govern read models of queries, which are plain views mapped inside each query and never aggregates.
