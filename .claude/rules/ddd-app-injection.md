---
paths:
  - "**/source/app/**"
---

# Application Injection Conventions

This governs how a use case gets what it works with: the repositories, queries and contracts a handler receives, who builds the handler, who owns the transaction, how process-lifetime objects are constructed, and how a handler records the events its entities raise. Where the files live is the layout rule's concern and how a use case is exposed is the entries rule's. Where it needs a language's own mechanism (file extension, visibility, how a type is made immutable or distinct, how an interface is marked), the language idioms rule for the file's extension supplies it.

---

## Dependency injection

- A handler picks its dependencies in the handler, not in a registry far from it: an `of` factory takes the object of repositories that the run's transaction provides and passes the handler its constructor arguments. That object exposes every repository, query, and shared writer (the audit log, the outbox) of the transaction, each typed by its core interface and a query by its concrete class, so a handler that asks for something that does not exist fails to compile in its own file. Adding a dependency is one edit in one file. The handler imports only the type of that object and of the queries from the persistence layer; it imports no injection library, because the project has none.
- There is no container, no token and no per-context registry: nothing is registered per use case or per repository, so adding a use case touches only its own folder and its context plugin, and adding a repository touches only the object that exposes the run's repositories.
- Objects that live as long as the process (the database client, the transactional boundary, the cache, the reporter, the mailer, the auth library and the session built over it) are built in one place, in a plain function, in the order they depend on each other, so the construction can be read top to bottom and a test replaces a part by passing it to that function. Nothing else constructs an infrastructure object.
- Each process's composition root, in the process entrypoint folder, calls that function and hands each context's plugin an object holding only what that context needs, typed by the core contracts it depends on rather than by the concrete classes, so the application layer never imports the layer that embeds the auth library. The plugin passes those values on to its use cases' plugins as concrete positional arguments.

## Handlers and the transaction

- A handler receives **only** the repositories, queries, and contracts it orchestrates, never the transactional boundary. The boundary builds the handler inside a run, from the repositories of that run's transaction, so every repository and query the handler receives already runs through that transaction. A handler that also received the boundary could only open a second, independent transaction, which breaks the atomicity of the entity write and its outbox event.
- A handler receives repositories and contracts through their core interfaces. A query has no interface: the handler types it as the concrete persistence class through a reference used only for its type, which, with the type of the repositories object, is the persistence import the application layer is allowed.
- Handlers are never cached or shared. A handler holds executor-bound repositories and queries, so keeping one would keep the transaction of the first run that built it, and that transaction is already closed by the next run; the boundary builds a new one for every run.
- The entry files, `plugin`, `processor`, `hook` and `reaction`, receive the transactional boundary as a factory argument and open the run themselves: the work they pass to the boundary builds the handler from the run's repositories and calls it with the parsed input. They take the boundary as an argument because the composition root builds it. A route behind the request gates receives its boundary already bound (to the confirmed tenant, or to the actor's identity for the steps that precede a tenant) and never builds a run context itself. The factory takes its dependencies as positional parameters while each has a distinct type, so two of them cannot be swapped silently, and declares no dependencies interface of its own.
- A use case that sends an email receives the mailer in its handler, never in its entry file, so the decision to send and what to send stay in the handler. The email leaves only after the commit, because a send can't be rolled back: that handler takes the shape of the several-commits exception below, with its one run building the run-scoped type and the send following it. A failed send is reported without failing the use case, since the state it announces is already committed.

## Exceptions

- **A write the handler retries after a constraint violation.** A run opened on the boundary that a run's own transaction provides is a savepoint of that same transaction, not a second one, so it is the only way to survive a violation without aborting the whole run. A handler that must retry one write (e.g. an insert whose generated unique value may already be taken) receives that savepoint boundary and uses it for nothing else.
- **A handler with no executor-bound repository or query.** A handler that orchestrates only non-transactional contracts, such as a cache or the session library, is built once by the factory that builds its entry and runs outside any run. Opening a database transaction for it would hold a pooled connection for the whole call while nothing writes through it. The moment such a handler gains a repository or query, it goes back to being built inside a run.
- **A use case that needs several commits.** Some use cases can't hold one transaction for their whole duration, such as a recomputation that marks an execution running, computes for a long time, then records the terminal state, or an invitation that is read under one context and settled under another. Its handler receives the transactional boundary instead of repositories and queries, is built outside any run, and opens one sequential run per commit. Each of those runs builds a run-scoped type declared beside the handler, in the handler's file and named `<UseCase>Steps`, which holds the repositories and queries and has its own `of` factory like any handler; the handler passes the work that builds it to the boundary. The runs are sequential and opened on the root boundary, never inside one another, because each must commit on its own. This and the retried write above are the only cases where a handler sees the boundary. A workflow handler follows the same shape: it receives the boundary and opens one run per context it calls.

## Events

- A handler that changes an entity takes the event the entity returned and appends it to its bounded context's outbox, through the outbox writer, in the same transactional run as the entity write. It never publishes to a queue, a bus, or any other system directly. A direct publish either announces a change that later rolls back or loses the event if the process dies right after the commit. The outbox makes the event commit with the change or not at all.
- A handler never calls the scheduler or cache contracts in reaction to its own write. Those effects belong to the event consumers, which receive the event from the outbox, so they happen after the commit and survive a crash.

---

## Non-goals

- Doesn't decide which dependencies a given use case needs — that follows from what the use case does.
- Doesn't govern the outbox relay — it is infrastructure, not an application-layer file.
