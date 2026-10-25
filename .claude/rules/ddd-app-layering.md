---
paths:
  - "**/source/app/**"
---

# Application Layout Conventions

This governs the internal folder and file layout of the application layer, which orchestrates use cases over the domain's bounded contexts: how use cases are grouped, which files a use case and its bounded context hold, and where flows that span several contexts live. How a handler gets its dependencies and its transaction is the injection rule's concern, and how a use case is exposed over HTTP, a queue or the auth library is the entries rule's. It doesn't cover the domain layer's own layout (that has its own rule) or any single use case's behavior (that belongs to its bounded context's spec). Where it needs a language's own mechanism (file extension, visibility, how a type is made immutable or distinct, how an interface is marked), the language idioms rule for the file's extension supplies it.

---

## Folder shape

- The application layer mirrors the domain's **bounded contexts**, but not their sub-slices: one folder per bounded context, named exactly as in core, so a use case's owning context is obvious from its path.
- The shared kernel is not mirrored. It is a domain concept and exists only in core, so the application layer holds nothing but real bounded contexts, their use cases, and the workflows that span them.
- Inside a bounded context, each use case gets its own folder, named verb-noun after what it does, joined the way the language's names allow (create a definition, not "definition service"). This avoids mirroring sub-slices because many use cases cross more than one entity of the same context, such as listing definitions with their last execution, or forcing a recomputation that loads a definition and creates an execution. A sub-slice folder would have no natural owner for them.
- Use-case folders are one level deep and never nest. A use case that seems to need internal structure is really two use cases.
- A use-case folder exists only for a use case with a handler that orchestrates injected repositories, queries, or contracts. Everything inside it feeds that one use case; the only things it receives from outside are its injected dependencies. Logic that only hardens or configures a third-party library (rejecting a request option, demanding re-authentication, patching a request body) has no handler and touches none of the project's own state, so it lives in the dedicated layer that embeds and configures that library, never here.
- An application context imports only its own folder, never another application context's implementation, so contexts never couple through each other's use cases or HTTP plumbing. The one exception is a workflow, below.

```
<context>/
├── <context>.plugin       # composes use-case plugins, route prefix, error → status
├── <context>.processor    # composes use-case processors, job name → processor
├── <context>.hooks        # gathers the use cases' auth-library entries
├── <context>.views        # output shapes shared by several use cases
└── <use-case>/
    ├── <use-case>.schema
    ├── <use-case>.handler
    ├── <use-case>.errors
    ├── <use-case>.plugin
    ├── <use-case>.processor
    ├── <use-case>.hook
    ├── <use-case>.mail
    └── <use-case>.mapper

<workflow>/                   # at the root, beside the context folders
├── <workflow>.schema
├── <workflow>.handler
├── <workflow>.reaction
└── <workflow>.errors
```

The tree writes the separator between prefix and kind as a dot; the language idioms rule says when it is another character.

## Workflows

- A flow that needs several bounded contexts is a workflow: its own folder at the root of the application layer, beside the context folders and named after the flow (verb-noun), never inside one of the contexts it touches. Otherwise that context would quietly own a rule about another one.
- A workflow is the only application code that may call the use cases of more than one context. Its handler calls each context's handlers in sequence, each in its own run, so a transaction still covers one context. A flow whose orchestration has invariants of its own (ordering, compensation after partial failure, idempotency across the boundary) also gets its own spec.
- A `reaction` is the entry the outbox relay calls for one event type of one context. It parses the event's payload with the workflow's input schema, translates the emitting context's ids into the receiving context's, and calls the workflow's handler. It carries no decision beyond that translation, so every decision stays in the handler.
- A reaction that only calls use cases of its own context is not a workflow: it is a job, a `processor` of that context. A consumer that only produces an infrastructure effect (rescheduling, cache overwrite) is infrastructure and never lives here.

## Files inside a use-case folder

Files are named `<use-case>.<kind>` plus the language's extension, with the prefix kept for the same reason as in core: tabs, search results, and stack traces identify the file without its path. Only the kinds a use case actually needs exist. A workflow folder uses the same prefix rule with the kinds `schema`, `handler`, `reaction` and `errors`.

| Kind | Holds |
|---|---|
| `schema` | The runtime validation schemas for the use case's input and output. Their types are derived from them and never declared by hand, so the runtime check and the type can't drift. |
| `handler` | The use-case class: it receives the parsed input, orchestrates repositories, queries, and contracts, and returns the output. It declares how it is built from the run's repositories in an `of` factory beside the constructor that receives them, and it is the entry service a transactional run builds. |
| `errors` | Orchestration errors that aren't domain rules, e.g. a requested entity that doesn't exist. Domain errors stay in core. |
| `plugin` | The HTTP-framework plugin that exposes this use case's route, reusing the same schemas for the request and response, or, for a request gate, the macro the routes declare. Present only for use cases reachable over HTTP or serving as a gate. |
| `processor` | The queue-framework job processor that invokes this use case from the worker, parsing the job payload with the same input schema. Present only for use cases a job triggers. |
| `hook` | The auth-library hook that enters this use case from the library: a request hook (before or after one of its endpoints), a database hook, or a library callback. Present only for use cases the auth library triggers. |
| `mail` | The email this use case sends, as a plain value of the project's mailer (recipient, template, and the variables the template fills), never one class per template. It belongs to the use case that sends it, never to a shared mail folder. Present only for use cases that send an email. |
| `mapper` | Pure functions that turn what an entry receives (a library route, a changed record) into the use case's input, or into nothing when the event doesn't concern it. Present only when that translation is more than a field copy, so it can be tested without the library. |

- An output shape that two or more use cases return (the same entity presented the same way) is declared once, in the context's views file, as a schema plus the function that maps the entity to it. The use cases' schemas and handlers reference it instead of repeating the fields, so the shape and its mapping can't drift between routes. A shape only one use case returns stays in that use case's schema.
- Schemas validate **shape only**: types, formats, presence, and conversion from wire values (e.g. an ISO string to a date). A business invariant is never re-checked in a schema, even if that would be easy, because it already lives in the entity. Two copies of one rule drift apart silently.
- A use case, like a domain piece, is as large as its flow needs. A handler that orchestrates many steps because the flow has many is essential complexity and stays one handler; splitting it only to shrink the file separates steps that change together.
- Tests don't live inside a use-case folder. Keeping them out keeps each folder limited to what ships.

---

## Non-goals

- Doesn't decide which use cases a bounded context exposes or what they guarantee — that's the bounded context's own spec.
- Doesn't govern the outbox relay or the infrastructure event consumers — those are infrastructure, not application-layer files.
