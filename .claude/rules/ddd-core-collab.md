---
paths:
  - "**/source/core/**"
  - "**/source/repos/**"
---

# Bounded Context Collaboration Conventions

This governs how bounded contexts work together without knowing each other: how one context refers to another, how a change in one triggers work in another, and how one reads what another owns. Where files live is the layout rule's concern and how each piece is written is the modeling rule's; this one is about the seams between contexts. The common thread is that every context stays understandable and testable alone, and all knowledge of more than one context sits in the application layer.

---

## References between contexts

- A context refers to another context's data by id only, declared as its own identifier value object under its own name. It never holds the other context's entity or imports its types, because the id is the one coupling that survives either side being rewritten.
- A context never imports another context, including for a type. The only shared vocabulary is the shared kernel's, so a type that two contexts both seem to need is either a shared-kernel primitive or two separate types, one per context.

## Reacting to a change: events by default

- A change in one context that must cause work in another travels as an event. The operation returns it, the caller commits it through the shared kernel's outbox contract in the same transaction, and a handler in the application layer then calls a use case of the receiving context. The emitting context never knows who listens, because that is what lets a listener be added or removed without editing the emitter.
- Consistency across contexts is eventual. A transaction covers one entity's change in one context, and a rule that seems to need two contexts in one transaction is a signal the boundary is drawn wrong, not a reason to widen the transaction.
- An event is the emitting context's public message: plain values only, never an entity, so a listener depends on the data and not on the emitter's internals. Only the application layer consumes it.

## Reading from another context: a contract in the consumer's language

- When a context needs information another one owns, it declares a role contract of its own (a reader, a resolver) in its own words and types. The application or infrastructure side implements it by calling the owning context. The consumer never learns the owner's model, so a change on the owner's side is absorbed by one adapter, the same role an anti-corruption layer plays.
- The contract returns only what the consumer's rule decides on, never the owner's entity, because a wider return type is how the owner's model leaks into the consumer.

## Orchestration lives in the application layer

- A flow that needs several contexts at once is coordinated by an application use case that calls each context through its barrel, in sequence, each in its own transaction. No context orchestrates another.
- A flow whose orchestration has invariants of its own (ordering, compensation after partial failure, idempotency across the boundary) is a named workflow with its own spec, never logic hidden inside one of the contexts it touches, because otherwise one context quietly owns a rule about another.

## The shared kernel is not a shortcut

- Only what every context needs enters the shared kernel. Two contexts wanting the same thing does not qualify, because a kernel grown by convenience becomes the hidden coupling the whole separation was meant to avoid. Until a third context needs it, the two contexts each keep their own.

---

## Non-goals

- Doesn't decide which contexts exist or which of them talk to each other; that belongs to each slice's spec.
- Doesn't govern the application or infrastructure layers' own layout, only what they may do between contexts.
