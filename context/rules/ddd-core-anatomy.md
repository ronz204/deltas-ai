---
paths:
  - "**/source/core/**"
  - "**/source/repos/**"
---

# Sub-slice Anatomy Conventions

This governs the files inside a sub-slice of a bounded context: the fixed set of file kinds, how repositories and role contracts are declared and named, how writes and reads are kept apart, and how persistence maps back to the domain. Where bounded contexts and sub-slices live, what core may depend on, and the barrel are in the layering rule; this file is only about what goes in a sub-slice's files. Where it needs a language's own mechanism, the language idioms rule for the file's extension supplies it.

---

## File naming inside a sub-slice

- Files are named `<subslice>.<kind>` plus the language's extension (the language idioms rule sets the separator between the two), where `<subslice>` matches the folder name, and `<kind>` is one of seven fixed kinds or a role, each answering a single question. A role kind (`notifier`, `scheduler`, a session writer) names a contract the domain needs from outside other than the entity's own persistence, so the kind is the role itself.

| Group | Kind | Holds |
|---|---|---|
| Model | `entity` | The root entity (the aggregate root), plus the state shape its reconstitution factory accepts |
| Model | `parts` | Entities owned by the root entity that aren't the root |
| Model | `vos` | Value objects, including the sub-slice's typed identifiers and closed sets of literal values (a state, a tier) |
| Behavior | `policies` | Pure domain rules that don't belong to a single entity: a decision computed from several values, a threshold policy, a domain service. No I/O; inputs in, verdict out |
| Signals | `events` | Domain events the root entity emits |
| Signals | `errors` | The domain error the sub-slice signals, with its closed set of codes |
| Boundary | `repository` | The entity's write-side persistence interface, plus the simple reads a write needs to decide: load one entity (by id, by natural key, or locked for update), count rows to enforce a limit, and insert or update, including compare-and-set transitions |
| Boundary | `<role>` (e.g. `notifier`, `scheduler`) | One contract the domain needs from outside that isn't the entity's persistence, one file per contract, with the role as the kind. This is what replaces a catch-all contracts file |

- There is no catch-all kind and no shared `contracts` file. A type that crosses the boundary goes in the file of the contract that carries it; a type internal to the model goes beside the model it describes. A catch-all file is where a type ends up whenever its author didn't decide which of those it is, and it hides that undecided boundary.
- One contract per file, named by its role, so a contract is found by its file name and a change to one never touches another. A repository or role file never re-exports a concrete entity, part or domain-error class just to make it reachable from outside — the barrel already does that.
- Persistence is split CQRS-lite between the entity's `repository` file and read-side queries. The test is the return type: a method that writes, or returns an entity, a count, or a boolean a write decides on, belongs to the repository; a read model for a reader is a query, which core does not declare. A query is a concrete class on the persistence side, one file per query named after what it reads with its full noun, and it never returns an entity. Its read model, whose name ends in `View`, is declared beside the query, except when a core policy decides on it: that one read model lives in core, in the file of the contract that serves it, in the sub-slice that owns the policy (or the shared kernel when several contexts use it). Keeping the two apart stops the repository from growing one bespoke finder per screen, and keeps read models free to join across tables without pretending to be entities.
- Method names follow fixed verbs, so a reader knows the shape of a call before opening it. A repository uses `find…` (load by id or natural key), `lock` (load for update), `insert`, `update`, `transition` (a compare-and-set; it returns whether it applied, or, when every caller would answer a stale transition with the same domain error, raises that error itself, so the check is written once), and `count…`. A name outside those verbs is a signal the method is on the wrong side: a read for a reader is a query, which exposes exactly one method, `execute`, because its file and class name already say what it reads.
- Only the kinds a sub-slice actually needs exist — a sub-slice with no domain events has no `events` file, and one needing nothing from outside has no `repository` or role file — because an empty placeholder communicates nothing a missing file doesn't already communicate.
- The `<subslice>.` prefix is kept even though the folder already disambiguates, so that editor tabs, search results, and stack traces identify a file unambiguously without its path.
- Files within a sub-slice import each other directly, never through the barrel. A sub-slice that needs something from the shared kernel imports it through that context's barrel, never by reaching into one of its sub-slice folders.

## Repositories and role contracts

- A `repository` or role file declares one contract the domain needs from outside, and lives in the sub-slice it serves. A contract every bounded context shares (such as the outbox writer) lives in the shared-kernel bounded context's own sub-slice. Every contract stays in core, including ones only the application layer calls (a scheduler, a lock, a cache), because a contract belongs to the side that depends on it. Declaring it in core keeps every dependency pointing inward.
- A contract interface's name marks it as an interface (by a prefix, a suffix, or nothing where the language names an interface by its behavior) and carries its role, matching the file it lives in; the language idioms rule says which. Its implementation is named so the two never collide. A persistence implementation takes no technology prefix, because the project has exactly one persistence technology and a prefix would only repeat it. An infrastructure implementation ends in `Adapter`. The contract and its implementation then never collide in a flat namespace, and an injection site reads as "depends on the interface", not on the concrete class.
- Mapping between a persisted row and a domain object belongs to the persistence layer, in a mapper beside the repository. The mapper serves the entity only; each query maps its own rows inside its own file, so changing one query's joins or columns never touches another. The mapper builds the entity's state with the same value-object parsers the domain uses, so a row that drifted from the domain's rules fails on read instead of flowing in. Columns that hold a closed set take their allowed values from the core literal sets, so the row type and the domain type are the same literal union.
- A contract is written in the domain's language (a report, a version, a fencing token), never a technology's (a connection, a queue option, a script). The test for where an interface belongs is who consumes it. If core or the application layer consumes it, it is a contract and gets its own file. If only adapters consume it and it speaks a technology's vocabulary, it is not a contract: it lives beside those adapters, and core never declares it. Declaring it in core would make core depend on a technology.
- When a contract's signature starts needing a technology's vocabulary, split it into a domain-language contract file in core and an adapter-internal interface in infrastructure, instead of letting the technology leak inward.

---

## Non-goals

- This rule does not decide where bounded contexts or sub-slices live or what core may import; that is the layering rule.
- It does not decide which domain concepts form a sub-slice or where an aggregate boundary ends; those are per-bounded-context decisions made in that slice's own `spec.md`.
