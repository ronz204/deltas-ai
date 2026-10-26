---
id: ddd-core
installs-as: code-domain-layer
applies-when: the project keeps a domain layer split into bounded contexts, with persistence code that maps to it
paths-default:
  - "**/source/core/**"
  - "**/source/repos/**"
---

# Domain Layer Conventions

This governs the domain layer (core) and the persistence code that maps to it: how bounded contexts are laid out and what they may import, the files inside a sub-slice, how each building block is written, and how contexts work together without knowing each other. The language idioms rule for the file's extension supplies every language mechanism (separators, visibility, barrel syntax, immutability, distinct types, interface marking). The application layer's layout, injection and entries are the application rule's.

---

## Bounded contexts

- One folder per bounded context directly under core, named as a two-word pair; the second word describes that context and is never a suffix repeated across all of them.
- A context imports only the shared kernel, never another context. Core never imports the application layer, infrastructure, persistence, runtime config or any framework or I/O library; what it needs from there it declares as a contract.
- The shared kernel has the same shape as any context. It holds what every context needs, even if another context owns the data behind it (tenant admission is the example), and nothing else: two contexts wanting the same thing is not enough, each keeps its own until a third needs it. A format value object several contexts share sits with the primitive factories, not in a sub-slice of its own.
- Every context uses sub-slice folders, even with one concept today. One barrel at the context root re-exports every sub-slice file; nothing outside imports from a sub-slice folder. The language's visibility gate decides what is public, not hand-filtering at the barrel. Exported names are unique across the context and carry their full domain noun (no bare `Id`).

## Sub-slices

- A sub-slice is a cohesive concept: typically one entity (the aggregate root) plus everything that exists only to serve it. Its folder is one short word from inside the context. Split on two concepts, never on size or technical layer; never nest.
- Files are `<subslice>.<kind>` inside `<context>/<subslice>/`, prefix always kept. Only the kinds needed exist. There is no catch-all or shared `contracts` file: a type goes in the file of the contract that carries it, or beside the model it describes. Files in a sub-slice import each other directly; the shared kernel is imported through its barrel only.

| Kind | Holds |
|---|---|
| `entity` | The root entity and the state shape its `restore` accepts |
| `parts` | Entities owned by the root |
| `vos` | Value objects, typed ids, closed literal sets |
| `policies` | Pure rules over several values or entities, no I/O: inputs in, verdict out |
| `events` | Domain events the root emits |
| `errors` | The sub-slice's one domain error and its closed code set |
| `repository` | The write-side persistence contract |
| `<role>` | One contract the domain needs from outside (`notifier`, `scheduler`), one file each |

## Contracts, repositories and queries

- A contract lives in core, in the sub-slice it serves (one every context shares, such as the outbox writer, in the shared kernel), even when only the application calls it. It is written in the domain's language, never a technology's; one only adapters consume, in a technology's vocabulary, lives beside those adapters, and a signature that starts needing that vocabulary is split in two. A repository or role file never re-exports a concrete class.
- Interfaces are marked and named by role per the language rule. A persistence implementation takes no technology prefix; an infrastructure one ends in `Adapter`.
- The repository is the write side: methods that write or return an entity, a count, or a boolean a write decides on. Verbs are fixed: `find…`, `lock`, `insert`, `update`, `transition` (compare-and-set: returns whether it applied, or raises the one stale-state error itself when every caller would), `count…`.
- A read for a reader is a query: a concrete persistence-side class, one file per query named with its full noun, one method `execute`, never returning an entity. Its read model (name ends in `View`) sits beside it, except one a core policy decides on, which lives in core with the contract serving it.
- The mapper beside a repository goes to the row and to the domain only and serves the entity only (each query maps its own rows). It parses every column with the domain's value-object parsers before `restore`; a repository never builds an entity or row by hand. Closed-set columns take their values from the core literal sets.

## Building blocks

- A rule lives with the entity or value object it constrains. Size alone is no smell; anything no domain rule requires (a wrapper, a variant of a base, extra indirection) is removed.
- Every piece is built from a shared-kernel base. A shape outside this table means extending the base once, never hand-rolling a variant in one sub-slice.

| Piece | Built from | Shape |
|---|---|---|
| Value object | Distinct-primitive support plus the text factory (or a hand-written parser) | Same-named declaration with `kind`, `parse`, `is`; type derived from it |
| Identifier | Identifier factory, one per id kind | As a value object; lowercase UUIDs |
| Literal set | Literal-set factory | `kind`, `values`, `parse`, `is`; type derived from `values` |
| Error | Base domain error's `define`, one type per sub-slice | Closed `snake_case` codes with fixed messages |
| Entity, part | Entity base typed by a state extending the base state | Frozen, `restore`, `sameAs`, protected `with`, intention-named factories |
| Event | Domain event type | `type`, `entityId`, `occurredAt`, `payload` |

- A value object's parser is its only constructor: it trims first, normalizes, validates and fails with the sub-slice's domain error, passing bounds as structured details. `is` accepts only already-normalized values. No class-based value object with a wrapped field. A literal set whose rejection is API contract passes its own error; otherwise the kernel's invalid-literal error is raised.
- An error carries a stable `snake_case` code (API contract); its code type is derived from the code map. Messages are fixed per code, in English, never interpolated; variable data travels in structured details.
- Entity state traits (version, timestamps, soft delete) are mixed in only when the table has those columns. The base assigns and freezes the state, so a class never repeats its field list. Construction is restricted and takes the whole state object, never positional arguments.
- Creation uses intention-named factories taking `(input, now)`; core never generates ids or reads the clock. A part is created only by an operation of its root.
- `restore` is inherited, trusts its input and is never rewritten. Identity is `sameAs`, never plain equality. No getters or snapshot method: read fields directly; a derived fact is a method named as a computation.
- A state change returns a new instance through the base `with` patch, which never carries the id; one that changes nothing may return the receiver. An event is named `<entity>.<past_tense_verb>` with plain values. An operation that raises events returns the change result (`next` plus `events`) and never publishes; one that raises none returns the bare instance.

## Collaboration between contexts

- A context refers to another's data by id only, as its own identifier value object, and never imports its types. A type two contexts both seem to need is a shared-kernel primitive or two separate types.
- A change that must cause work elsewhere travels as an event of plain values: the operation returns it, the caller commits it through the outbox contract in the same transaction, and the application layer reacts. The emitter never knows its listeners. Consistency across contexts is eventual; needing two contexts in one transaction means the boundary is wrong.
- To read what another context owns, the consumer declares its own role contract (a reader, a resolver) in its own words, returning only what its rule decides on, never the owner's entity. The application or infrastructure side implements it by calling the owner.
- A flow that needs several contexts is an application-layer workflow, never logic inside one of them.

---

## Non-goals

- Does not decide which contexts, sub-slices or aggregate boundaries exist, nor which contexts talk; that is each slice's spec.
- Does not govern the application or infrastructure layers' layout.

<!-- rationale -->

## Why the rule is shaped this way

Bounded contexts
- A shared two-word shape keeps the set of contexts scannable as a flat list. A suffix repeated on every folder (`-domain`, `-core`) stops carrying information once everyone has it.
- Contexts importing each other end in import cycles and coupling through internals, so neither can be understood or tested alone. Core importing one adapter-side module means every change on that side can break the domain; the dependency direction keeps the domain replaceable around its adapters.
- A kernel grown by convenience becomes the hidden coupling the separation was meant to avoid. What qualifies: admission to a tenant (membership standing, required second factor, permission check) runs in every context's endpoints, so its policy and the read model it decides on are kernel, while the context owning the membership tables implements the reads and the use case that runs it. A sub-slice holding one shared format value object (an email address) and its error is a folder with nothing to group.
- Deciding folder shape by how many concepts a context has today is an ad hoc call that drifts silently as it grows.
- The barrel is the only way in, otherwise other contexts couple to file layout that is free to change. Using visibility as the gate avoids a second hand-kept filter. Folder names are invisible to importers, hence full domain nouns.

Sub-slices and files
- Splitting by layer scatters one concept across two trees, so each change touches both. A big entity with many policies is essential complexity of a rich domain; splitting it to shrink the folder separates rules that change together. Nesting means two sub-slices side by side.
- The `<subslice>.` prefix stays although the folder disambiguates, so tabs, search results and stack traces identify a file without its path. An empty placeholder says nothing a missing file doesn't.
- A catch-all file is where a type lands when its author did not decide whether it is boundary or model; it hides that undecided boundary. One contract per file means a change to one never touches another, and re-exporting a concrete class duplicates what the barrel already does.

Contracts and persistence
- A contract belongs to the side that depends on it, so declaring it in core keeps every dependency pointing inward; a technology-shaped interface in core would make core depend on that technology. Naming the persistence implementation without a prefix is fine because the project has exactly one persistence technology; `Adapter` keeps the two from colliding in a flat namespace and makes an injection site read as depending on the interface.
- The repository/query split (CQRS-lite) stops the repository growing one bespoke finder per screen and leaves read models free to join across tables without pretending to be entities. Fixed verbs tell a reader a call's shape before opening it; a name outside them is a read on the wrong side.
- Parsing rows with the domain's own parsers makes a row that drifted from the domain's rules fail on read instead of flowing in. Sharing the literal sets makes the row type and the domain type the same union. Each query mapping its own rows means changing one query's joins never touches another.

Building blocks
- Scattering a rule across files makes every change to it touch all of them. Splitting or generalizing a complex entity to look smaller adds indirection that means nothing in the domain. A hand-rolled variant is how two entities end up with two ways to rebuild from storage.
- The parser as only constructor keeps two string-backed concepts from being swapped silently while equality stays plain and the value maps to a column unwrapped; deriving the type from the declaration means it cannot drift from its parser. A wrapped-field class adds an unwrap at every boundary. `is` that accepts unnormalized values would report true for something `parse` changes.
- Stable codes let the HTTP layer map a code to a status without knowing classes, so renaming a class or rewording a message never breaks a client; a status map keyed by the derived code type fails to compile when a code has no status. Interpolated messages split one failure into many issues in monitoring and leak user-supplied values into logs.
- A trait without a column is a field the mapper must invent. Positional state arguments of the same type swap silently. Core generating ids or reading the clock makes factories nondeterministic. Two restores of one row are different objects but the same entity, hence `sameAs`.
- Events only commit with the change: the entity cannot know whether its change commits, so only the caller, which owns the transaction, hands events to the outbox. Returning a bare instance when no events exist makes the signature say whether events exist.

Collaboration
- An id is the one coupling that survives either side being rewritten. Events let a listener be added or removed without editing the emitter. A wider return type is how the owner's model leaks into the consumer; the role contract plays the part of an anti-corruption layer, absorbing owner changes in one adapter.
- A flow living inside one context quietly makes that context own a rule about another.
