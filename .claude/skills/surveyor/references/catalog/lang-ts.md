---
id: lang-ts
installs-as: code-ts-idioms
applies-when: the project's source is TypeScript under a DDD layout
paths-default:
  - "**/source/**/*.ts"
---

# TypeScript Idioms Conventions

This governs how TypeScript expresses what the DDD rules state in neutral terms: file naming, the public surface of a bounded context, distinct primitive types, immutability, construction, interfaces, errors, reuse of a base and factories. The DDD rules defer here for each answer; this rule does not repeat what they decide.

---

## Names

- Files are `<subslice>.<kind>.ts` and `<use-case>.<kind>.ts`; a dot separates the prefix from the kind.
- Folders are kebab-case, so a two-word context or a verb-noun use case reads as one name.
- Identifiers use `camelCase` for functions, methods and fields, `PascalCase` for types and classes. The base's operations (`restore`, `sameAs`, `with`) and the event's fields (`entityId`, `occurredAt`) follow it.

## Public surface

- A context's entry point is an `index.ts` at its root with one `export * from "./<subslice>/<file>"` per file that exports a symbol.
- `export` in a concrete file is the public/private gate: an internal symbol is simply not exported. `index.ts` changes only when a file is added or removed.
- Files in a sub-slice import each other by relative path and reach into the shared kernel only through its barrel.

## Types and construction

- A value object is a branded primitive: a brand type plus a same-named constant holding `kind`, `parse` and `is`. The type is inferred from the constant (`typeof`/`ReturnType`), never declared by hand. Equality stays `===`.
- State is a named `interface` extending the base entity state, with `readonly` fields; the base freezes the instance.
- The constructor is `private` and takes the whole state object. Creation goes through `static` factories; `restore` is inherited; a state change goes through the base's `protected with`.
- A handler's `of` is a `static` method beside its constructor. A run-scoped class beside a multi-commit handler is a plain class named `<UseCase>Steps` in the handler file.

## Interfaces, errors and bases

- A contract is an `interface` named `I` plus its role (`IOrderRepository`, `INotifier`). Its implementation drops the `I`, or ends in `Adapter` for infrastructure.
- A domain error is a class extending the base domain error's `define` result; a failure is a `throw`, and the application layer maps its `code`.
- Entities and parts extend the shared kernel's entity base class. Its factories (text value object, identifier, literal set, error `define`) are plain functions.
- A reference used only for its type is an `import type`, which is how a handler names a query's concrete class and the repositories object without runtime imports from persistence.

---

## Non-goals

- Does not restate what the DDD rules decide, only how TypeScript spells it.
- Does not govern formatting or lint style, which the project's tooling owns.

<!-- rationale -->

## Why the rule is shaped this way

- The dot separator works because TypeScript module names allow it, and the prefix keeps tabs and stack traces unambiguous. Kebab-case folders keep multi-word names readable in a path.
- Once a file exports something it flows through the barrel automatically, so the visibility keyword is the only gate and there is no second list to keep in sync.
- Inferring the brand type from the constant means the kind is declared once and the type cannot drift from its parser; `===` and unwrapped column mapping come free with a branded primitive.
- A private constructor plus a whole-state argument stops direct construction and positional swaps; freezing makes immutability true at runtime as well as in the type system.
- The `I` prefix plus an unprefixed implementation or an `Adapter` suffix keeps contract and implementation from colliding in a flat namespace, and injection sites read as depending on the interface.
- `import type` erases the import, so application code can name persistence types without loading persistence at runtime.
