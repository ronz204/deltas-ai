# Checkout — Spec

> Turns a cart into a paid, persisted order by orchestrating pricing, stock, and payment.

## Intent

A customer is charged exactly once, for the right amount, only when the goods are held for them, and an order exists only if they paid.

## Scope

Owns the order flow from request to saved order. Does not compute money, hold stock, or talk to a payment provider itself: it delegates to pricing, stock, and an injected payment gateway and order store.

## Contract

| Item | Behavior |
|---|---|
| `createCheckout({ stock, gateway, orders })` | Returns a `checkout(request)` function. |
| `checkout({ orderId, lines, percentOff?, taxBps })` | Resolves to the saved `Order`. |
| Errors | `EmptyCartError`, `DuplicateOrderError`, `PaymentFailedError`; stock errors pass through. |

## Invariants

1. Checkout of an empty cart throws `EmptyCartError` before touching stock or payment.
2. Stock is reserved before payment is attempted, and the reservation is released if payment fails.
3. Payment is attempted at most once per `orderId`; a repeated checkout for the same order throws `DuplicateOrderError`.
4. Order totals come only from the pricing slice: checkout performs no money arithmetic of its own.
5. An order is persisted only after payment succeeds.

## Deferred / Open questions

- Partial payments and refunds are out of scope until a payment provider is chosen.

## Acceptance criteria

- `tests/checkout.test.ts` passes.

---

Last updated: 2026-10-25.
