---
id: lang-py
installs-as: code-py-idioms
applies-when: the project's source is Python under a DDD layout
paths-default:
  - "**/source/**/*.py"
---

# Python Idioms Conventions

This governs how Python expresses what the DDD rules state in neutral terms: file naming, the public surface of a bounded context, distinct primitive types, immutability, construction, interfaces, errors, reuse of a base and factories, plus how Python projects here are set up. The DDD rules defer here for each answer; this rule does not repeat what they decide.

---

## Project setup

- `uv` manages the environment and dependencies, `ruff` lints and formats, both configured in `pyproject.toml`.
- Pydantic is the library for validated data: schemas, value objects and entity state. It is a deliberate dependency of core.
- Packages have no `__init__.py`; implicit namespace packages make it unnecessary. Code that would live in one goes in the context's public module.

## Names

- Modules and folders are `snake_case`. Files are `<subslice>_<kind>.py` and `<use-case>_<kind>.py`, because a dot would read as a package boundary. A two-word context is `sales_orders`, a verb-noun use case `create_order`.
- Identifiers use `snake_case` for functions, methods and fields, `PascalCase` for classes: `restore`, `same_as`, `entity_id`, `occurred_at`. The copy-with-changes operation is `with_`, because `with` is reserved.

## Public surface

- A context's entry point is an `api.py` at its root that imports the public names of every sub-slice module explicitly and lists them in `__all__`. Never star imports.
- A new public symbol is one edit in `api.py`. An internal name starts with an underscore and is not listed.
- Files in a sub-slice import each other by relative or absolute module path and reach into the shared kernel only through its `api.py`.

## Types and construction

- A value object is a `NewType` over its primitive paired with a validating pydantic type (an `Annotated` type applied through a `TypeAdapter`) as its parser. The kernel's factories build both from one declaration. Equality stays `==`.
- State is a pydantic model with `frozen=True` extending the base entity state.
- Python cannot make a constructor private, so an entity is built only through its `classmethod` factories, never by calling the class in application code. `restore` is a `classmethod` on the base built over `model_construct`; a state change goes through the base's `with_`, over `model_copy(update=...)`.
- A handler's `of` is a `classmethod` beside its initializer. A run-scoped class beside a multi-commit handler is a plain class named `<UseCase>Steps` in the handler module.

## Interfaces, errors and bases

- A contract is a `typing.Protocol` named `I` plus its role (`IOrderRepository`, `INotifier`). Its implementation drops the `I`, or ends in `Adapter` for infrastructure.
- A domain error is an exception class deriving from the base domain error's `define` result; a failure is a `raise`, and the application layer maps its `code`.
- Entities and parts extend the shared kernel's entity base, a frozen pydantic model. Its factories (text value object, identifier, literal set, error `define`) are plain functions.
- A reference used only for its type goes under `if TYPE_CHECKING:`, which is how a handler names a query's concrete class and the repositories object without runtime imports from persistence.

---

## Non-goals

- Does not restate what the DDD rules decide, only how Python spells it.
- Does not govern which lint rules `ruff` enables, which `pyproject.toml` owns.

<!-- rationale -->

## Why the rule is shaped this way

- `uv` and `ruff` settle environment and style so neither is argued in review. An empty `__init__.py` adds nothing a reader or a tool needs.
- Underscore file names are forced by the import system; hyphens cannot be imported.
- Explicit imports with `__all__` keep the surface a deliberate list instead of whatever a module defines, which star imports hide and `ruff`'s pyflakes rules flag. Python has no `export` gate, hence the underscore convention.
- A `NewType` plus a pydantic parser keeps two string-backed concepts from being swapped silently while equality and column mapping stay plain.
- `model_construct` trusts its input, which is correct because the mapper already parsed every field. Restricting construction to factories is convention only, since the language cannot enforce it.
- The `I` prefix and unprefixed or `Adapter` implementation avoid name collisions in a flat namespace and make injection sites read as depending on the interface.
- `TYPE_CHECKING` imports are erased at runtime, so application code can name persistence types without loading persistence.
