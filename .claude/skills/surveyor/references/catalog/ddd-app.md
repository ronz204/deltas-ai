---
id: ddd-app
installs-as: code-app-layer
applies-when: an application layer orchestrates use cases over the domain's bounded contexts
paths-default:
  - "**/source/app/**"
---

# Application Layer Conventions

This governs the application layer, which orchestrates use cases over the domain's bounded contexts: its layout, how a use case is exposed over HTTP, a queue or the auth library, and how a handler gets its dependencies and its transaction. The domain layer's rules and any single use case's behavior (its context's spec) are out of scope. The language idioms rule for the file's extension supplies every language mechanism.

---

## Layout

- One folder per bounded context, named as in core. The shared kernel is not mirrored, and neither are sub-slices: use cases often cross entities of one context.
- Inside a context, one folder per use case, verb-noun, one level deep, never nested (a use case needing structure is two). It exists only for a handler that orchestrates injected repositories, queries or contracts; logic that only configures a third-party library has no handler and lives in the layer that embeds the library.
- A context imports only its own folder, never another's implementation. The one exception is a workflow.
- A flow needing several contexts is a workflow: its own folder at the root of the layer, verb-noun, never inside a context. Its handler calls each context's handlers in sequence, each in its own run. Invariants of its own (ordering, compensation, idempotency) earn its own spec.
- A `reaction` is the entry the outbox relay calls for one event type of one context: it parses the payload with the workflow's input schema, translates ids between contexts and calls the handler, deciding nothing. One that only calls its own context is a job (`processor`); a consumer with only an infrastructure effect (rescheduling, cache overwrite) is infrastructure.
- Files are `<use-case>.<kind>` (a workflow uses `schema`, `handler`, `reaction`, `errors`), prefix kept, only kinds needed. Tests never live in a use-case folder.
- Per context: `<context>.plugin`, `.processor` and `.hooks` compose the use cases' entries, and `<context>.views` declares an output shape (schema plus entity mapper) that two or more use cases return.

| Kind | Holds |
|---|---|
| `schema` | Input and output runtime schemas; types derived from them, never hand-declared. Shape only: never re-check a business invariant |
| `handler` | The use-case class; its `of` factory beside the constructor builds it from the run's repositories |
| `errors` | Orchestration errors that aren't domain rules (entity not found); domain errors stay in core |
| `plugin` | The HTTP plugin exposing the route with the same schemas, or the macro of a request gate |
| `processor` | The queue job processor; parses the payload with the same input schema |
| `hook` | The auth-library entry: a request hook, a database hook or a callback |
| `mail` | The email as a plain value of the project's mailer, owned by the use case that sends it |
| `mapper` | Pure functions from what an entry receives to the use case input (or nothing); only when more than a field copy |

## Entries

- An entry only translates what a framework gives it into the use case's input and calls the handler; every decision stays in the handler.
- A context has one plugin that composes its use cases' plugins under one route prefix and maps domain and orchestration errors to status in one place; the api entrypoint mounts only these.
- A context with job-triggered use cases has one processor routing job name to use case and mapping errors to retry semantics in one place (retry transient failures, fail permanently on domain or orchestration errors). A context with library-triggered use cases has one hooks file building the objects the auth library accepts and mapping errors to the library's error type, typed with the library's own types, never an infrastructure type.
- The HTTP framework is imported only in plugin files, the queue framework only in processors, the auth library only in hooks. This layer never imports the layer that embeds the auth library.
- A context in this layer has no barrel: each entrypoint imports the per-context file for its own framework directly.
- A `processor` enters a use case by job (cron tick, queued recomputation), never as a reaction to a domain event. A `hook` checks the route it reacts to, maps the library context to the input schema and calls the handler; when its caller must not see a failure (it runs after the library committed) it validates itself, opens the run and reports without rethrowing.
- Prefer the library's extension points (database hooks, which run after commit) to matching routes. An audit entry is never such a reaction: the audit log records only actions the project writes itself, in the same transaction.
- Each guard, hook or mapper declares the library endpoints it reacts to as constants beside its own code.
- The request gate splits by what it decides. Requiring a session belongs to the infrastructure auth plugin (session typed as the core contract, never the library). Admitting a request into its tenant is a use case of the context owning membership: its handler fetches through queries and a kernel policy decides, returning a tagged outcome (admitted with access, or denied) instead of throwing. Its plugin registers the gate macro, maps a denial to a response and hands the route a transactor already bound to the tenant. The composition root builds the auth plugin once, builds the gate from it and passes the gate to each context's plugin; the library's own routes are mounted once at the entrypoint.

## Injection and transactions

- A handler picks its own dependencies: `of` takes the object of repositories the run's transaction provides (every repository, query and shared writer such as the audit log and outbox, typed by core interface, a query by its concrete class). There is no container, token or per-context registry: adding a use case touches its folder and its context plugin, adding a repository touches only that object. The handler imports only the types of that object and of queries from persistence.
- A handler receives only repositories, queries and contracts, never the transactional boundary. The boundary builds it inside a run, new for every run and never cached. A query is typed as the concrete class through a type-only reference.
- Entries (`plugin`, `processor`, `hook`, `reaction`) receive the boundary as a factory argument and open the run; a gated route receives it already bound. Dependencies are positional while each has a distinct type, with no interface of their own.
- Process-lifetime objects (database client, boundary, cache, reporter, mailer, auth library, session) are built in one plain function in dependency order; nothing else constructs infrastructure. Each process's composition root calls it and hands each context's plugin only what it needs, typed by core contracts.
- A handler that sends email receives the mailer itself, not its entry. The mail leaves only after the commit, and a failed send is reported without failing the use case.
- Exceptions, the only cases a handler sees the boundary or runs outside a run: a write retried after a constraint violation receives a savepoint boundary, used for nothing else; a handler with no executor-bound repository or query (cache, session library) is built once and runs outside any run, returning to a run when it gains one; a use case needing several commits receives the boundary, is built outside any run and opens one sequential run per commit on the root boundary, each building a `<UseCase>Steps` type declared in the handler's file with its own `of`; a workflow handler does the same, one run per context.
- A handler that changes an entity appends the returned event to its context's outbox in the same run as the write. It never publishes to a queue or bus, and never calls scheduler or cache contracts for its own write: those are event consumers' work after the commit.

---

## Non-goals

- Does not decide which use cases, routes, jobs or library entries a context exposes; that is the context's spec.
- Does not govern the outbox relay, event consumers or the layer that embeds the auth library; those are infrastructure.

<!-- rationale -->

## Why the rule is shaped this way

Layout
- Mirroring sub-slices would leave no natural owner for use cases that cross entities of one context, such as listing definitions with their last execution or a recomputation that loads one entity and creates another. The kernel is a domain concept and exists only in core.
- A workflow inside one context would make that context quietly own a rule about another. Contexts importing each other's use cases or plumbing couple through them.
- The prefix stays for the same reason as in core: tabs, search results and stack traces identify the file without its path. Tests stay out so each folder holds only what ships.
- Schemas validating shape only avoid two copies of one rule drifting apart silently; the entity already owns the invariant. A handler as large as its flow needs is essential complexity, and splitting it only to shrink the file separates steps that change together.
- A shape shared by two use cases declared once in views cannot drift between routes.

Entries
- One error-to-status translation per context avoids repeating it per route and letting identical errors drift to different statuses. Adding a use case then touches its own context and nothing outside.
- Both processes run the same handlers, so each must load only its own frameworks; a barrel would have to leave out every framework-bound file and end up exporting only what nothing outside imports.
- Typing hooks with the library's own types keeps the application layer from importing the layer that embeds the library; that layer accepts these entries through a shape of its own and only the composition root sees both sides.
- Database hooks run after the library's commit and carry the endpoint context, so reconstructing state by matching routes and capturing identity before the endpoint runs is avoided. Audit entries must be in the same transaction as the action, which a library-side reaction cannot give.
- Declaring endpoints beside the code that compares against them keeps each piece's dependencies in its own file; a library rename then means searching for an endpoint string in the few files that react to it.
- The gate keeps library logic (session) apart from tenant admission, which is a project rule owned by the membership data. Returning a tagged outcome avoids signalling denial by failure, and a pre-bound transactor means a gated route never receives the root one.

Injection
- Picking dependencies in the handler means a missing dependency fails to compile in its own file, and adding one is one edit in one file. A registry far from the handler adds a place to forget.
- A handler that also received the boundary could only open a second, independent transaction, breaking the atomicity of the entity write and its outbox event. A cached handler would keep the transaction of the first run that built it, already closed by the next.
- Positional dependencies cannot be swapped silently while each has a distinct type. Building infrastructure in one function, in order, lets a test replace a part by passing it.
- Mail cannot be rolled back, so it leaves after the commit; the state it announces is already committed, so a failed send does not fail the use case.
- A run opened on a run's own boundary is a savepoint of the same transaction, the only way to survive a constraint violation without aborting the run. A cache-only handler outside a run avoids holding a pooled connection while nothing writes through it. A long computation cannot hold one transaction, so each commit gets its own run, and runs are sequential because each must commit alone.
- A direct publish either announces a change that later rolls back or loses the event if the process dies after the commit; the outbox makes the event commit with the change or not at all. Effects reacting to a write belong to consumers so they happen after the commit and survive a crash.
