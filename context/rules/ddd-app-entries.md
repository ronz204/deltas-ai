---
paths:
  - "**/source/app/**"
---

# Application Entries Conventions

This governs how a use case is reached from outside: HTTP plugins, queue processors, auth-library hooks, the request gate that admits a request into its tenant, and the public surface of an application context. An entry only translates what a framework gives it into a use case's input and calls the handler; every decision stays in the handler. Where the files live is the layout rule's concern and how a handler gets its dependencies and transaction is the injection rule's. Where it needs a language's own mechanism (file extension, visibility, how a type is made immutable or distinct, how an interface is marked), the language idioms rule for the file's extension supplies it.

---

## Per-context entries

- Each bounded context has one plugin of its own that composes its use cases' plugins and sets their shared route prefix. The api entrypoint mounts only these per-context plugins, never a use case's plugin directly, so adding a use case touches its own context and nothing outside it.
- The per-context plugin maps errors to HTTP status codes in one place, both domain errors and the context's orchestration errors. Mapping them per use case would repeat the same translation in every route and let identical errors drift to different statuses.
- Each bounded context that has job-triggered use cases has one processor of its own. It routes each job name to its use case's processor and maps errors to the queue's retry semantics in one place: retry a transient failure, fail permanently on a domain or orchestration error that a retry can't fix. The worker entrypoint registers only these per-context processors.
- Each bounded context that has library-triggered use cases has one hooks file of its own. It builds the objects the auth library accepts (its database hooks, its callbacks, and any request middleware a use case needs) by calling its use cases' hooks in sequence, and maps domain and orchestration errors to the library's error type in one place. It types what it returns with the library's own types, never with an infrastructure type, so the application layer never imports the layer that embeds the library. That layer accepts these entries through a shape of its own that they satisfy structurally, composes them with its own guards, and plugs them in as they are; only the composition root sees both sides.
- The HTTP framework is imported **only** in plugin files, the queue framework **only** in processor files, and the auth library **only** in hook files. Handlers, schemas, and errors depend on none of them, because both processes run the same handlers and each one must load only its own frameworks.
- A bounded context in the application layer has no barrel. Its public surface is its per-context files (the plugin, the processor, the hooks), each imported directly by the one composition root or entrypoint that needs it: each entrypoint takes its framework's file. A barrel would have to leave out every framework-bound file to keep a process from loading a framework it does not use, and would be left exporting only what nothing outside the context imports. Per-context plugins, processors, and hooks files import their use cases' entry files by relative path.

## Kinds of entry

- A `processor` is an entry into a use case, not a reaction to a domain event. A job that runs a use case (a cron tick or a queued recomputation) gets a `processor` beside that use case. An event that must reach another context is a workflow's `reaction`, and one that only produces an infrastructure effect (rescheduling, cache overwrite) is infrastructure, never a processor.
- A `hook` is an entry too, for requests the auth library serves itself: it checks the route it reacts to, maps the library's context to the use case's input schema, and calls the handler. It carries no logic beyond that translation, so every decision stays in the handler, testable without the library. When its caller must not see the use case fail, such as a reaction that runs after the library already committed, it also validates the input itself, opens the run, and reports a failure without rethrowing. Reading what the library already exposes on its context (the session, its credential lookup, its password hasher) counts as translation, so the handler receives the result instead of a contract that wraps the library.
- Prefer the library's own extension points to reconstructing state: a use case that reacts to a change the library persists itself enters from its database hooks, which run after its commit and carry the endpoint context, not by matching routes and capturing identity before the endpoint runs. Audit entries are never one of these reactions: the audit log only records actions the project writes itself, in the same transaction.
- Each guard, hook, or mapper declares the auth library's endpoints it reacts to itself, as constants beside the code that compares against them, so everything a piece depends on lives in its own file. A library rename then means searching for the endpoint string, which is acceptable because each endpoint appears in only the few files that react to it.

## The request gate

- A gate every context's routes need splits by what it decides. What belongs to the auth library alone, such as requiring a session, lives in the infrastructure auth plugin. It receives the session contract typed as the core session contract and never as the library, and hands it to routes through the request context, together with a boundary bound to the actor's identity for the use cases that run before a tenant is known.
- Admitting a request into its tenant is not library logic: it is a use case of the bounded context that owns the membership data. Its handler fetches what the decision needs through queries and leaves the decision itself to a shared-kernel policy in core, which returns a tagged outcome (admitted with its access, or not admitted with its denial) instead of signalling a failure.
- Its plugin registers the gate's macro, opens the run, maps a denial to a response, and hands the route a transactor already bound to the confirmed tenant, so a route behind the gate never receives the root transactor.
- The composition root builds the auth plugin once, builds the owning context's gate from it, and passes that gate to each context's plugin, which uses it to get the gates. The library's own routes are mounted separately, once, at the entrypoint, because a mount inside the shared plugin would be repeated under every context's prefix.
- An application context imports only its own folder, the auth plugin's type, and the type of the gate it receives, so contexts never couple through each other's HTTP plumbing.
- The application layer never imports the layer that embeds the auth library, in any file kind. What it needs from the session it receives as a core contract, through the request context the auth plugin fills.

---

## Non-goals

- Doesn't decide which routes, jobs or library entries a context exposes — that's the bounded context's own spec.
- Doesn't govern the layer that embeds the auth library or the outbox relay — those are infrastructure.
