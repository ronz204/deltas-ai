---
paths:
  - "**/source/**/*.ts"
---

# TypeScript Idioms Conventions

This governs how TypeScript expresses what the DDD rules state in neutral terms: file naming, the public surface of a bounded context, distinct primitive types, immutability, construction, interfaces, errors, reuse of a base and factories. Each section answers one expression point, and the DDD rule that needs it defers here for the answer. It doesn't repeat what the DDD rules decide, only how TypeScript spells it.

---

## Names

- Files are `<subslice>.<kind>.ts` and `<use-case>.<kind>.ts`: a dot separates the prefix from the kind, because TypeScript module names allow it and the prefix keeps tabs and stack traces unambiguous.
- Folders are kebab-case, so a two-word bounded context or a verb-noun use case reads as one name in a path.
- Identifiers use the language's usual case: `camelCase` for functions, methods and fields, `PascalCase` for types and classes. The base's operations (`restore`, `sameAs`, `with`) and the event's fields (`entityId`, `occurredAt`) follow it.

## Public surface

- A bounded context's public entry point is an `index.ts` at its root that re-exports every file of every sub-slice with `export * from "./<subslice>/<file>"`, one line per file that exports at least one symbol.
- The `export` keyword in a concrete file is the public/private gate. A symbol that must stay internal simply isn't exported from its own file, and once a file exports something it flows through the barrel automatically, so `index.ts` changes only when a file is added or removed.
- Files within a sub-slice import each other by relative path, and reach into the shared kernel only through its barrel.

## Distinct primitives

- A value object is a branded primitive: a brand type over the primitive, plus a same-named constant that holds `kind`, `parse` and `is`. The type is inferred from the constant (`typeof`/`ReturnType`), never declared by hand, so it can't drift from its parser.
- Equality on a value object stays `===`, and it maps to a column with no wrapping.

## Immutability and construction

- State is a named `interface` extending the base entity state, with fields `readonly`. The base freezes the instance, so the fields are immutable in the type system and at runtime.
- The constructor is `private` and takes the whole state object. Creation goes through `static` factories, and `restore` is inherited from the base.
- A state change goes through the base's `protected with` patch, which returns a new instance.

## Interfaces

- A contract is an `interface` whose name starts with `I` followed by its role (`IOrderRepository`, `INotifier`). Its implementation drops the `I`, or ends in `Adapter` when it is an infrastructure implementation, so the two never collide in a flat namespace.

## Errors

- A domain error is a class extending the base domain error's `define` result. A failure is signalled by `throw`, and the application layer maps the error's `code` to a status or retry behavior.

## Reuse of a base

- Entities and parts extend the shared kernel's entity base class. The shared kernel's factories (text value object, identifier, literal set, error `define`) are plain functions.

## Factories and imports

- A handler's `of` is a `static` method beside its constructor.
- A reference used only for its type is written as `import type`, which is how a handler names a query's concrete class and the repositories object without importing runtime code from the persistence layer.
- A run-scoped class declared beside a multi-commit handler is a plain class in the handler file named `<UseCase>Steps`.

---

## Non-goals

- Doesn't restate what the DDD rules decide, only how TypeScript expresses it.
- Doesn't govern formatting or lint style, which the project's own tooling owns.
