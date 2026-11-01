# Specs and designs across services

Loaded by archivist when the repo is genuinely split into multiple independently-deployable services (for example a backend and a separate frontend, each with its own root) and a slice's files have to be placed. A project that merely has several top-level directories does not qualify: confirm the split is real before leaving the root `deltas/` default.

`deltas/` moves down one level, into each service's own root. A slice's files don't have to live under the same service: each lives under whichever service owns that layer. `<slice>.design.md` goes under the service that renders the UI, `<slice>.spec.md` under the service that implements the logic it describes. The slice name is the same and the `deltas/` directories differ. Don't co-locate them just because they share a name.

Since the files are no longer side by side, each must name the other service(s) it pairs with (a service name is durable architecture vocabulary, not a volatile reference). This runs both ways: `design.md` names which service's `spec.md` it renders, and that `spec.md` notes where its `design.md` and any sibling `spec.md` live, so either file is a complete map of the slice on its own.

A slice can also have more than one `<slice>.spec.md`, one per service, each scoped to the logic that service owns. This is not duplication: a backend spec might own session validity and data rules while a frontend spec for the same slice owns client-side navigation. Each such spec states plainly what it owns versus what it only reacts to from its siblings, so a reader of either knows which decisions are made there and which elsewhere.
