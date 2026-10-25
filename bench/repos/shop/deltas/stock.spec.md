# Stock — Spec

> Tracks on-hand quantity per SKU and short-lived reservations against it.

## Intent

Orders must never promise stock that is not there, and a customer who walks away must not hold stock forever. Stock owns that bookkeeping and tells the rest of the system when it changes.

## Scope

Owns on-hand quantities, reservations, and the `StockChanged` event. Does not know about prices, payment, or orders; it is a dependency of checkout, never the other way around.

## Contract

| Function | Behavior |
|---|---|
| `receive(sku, qty)` | Adds `qty` (a positive integer) to on-hand. |
| `available(sku)` | Returns on-hand minus the quantity held by live reservations. |
| `reserve(orderId, sku, qty)` | Holds `qty` for the order, or throws `InsufficientStockError`. |
| `release(orderId)` | Drops every reservation held by the order. |
| `commit(orderId)` | Consumes the order's reservations from on-hand. |

## Invariants

1. Reservations are idempotent per order and SKU: reserving again for the same `orderId` and `sku` changes nothing.
2. Available quantity is on hand minus reserved, and a reservation larger than the available quantity is rejected, so available never goes negative.
3. Every change to on-hand or reserved quantities (receive, reserve, release, commit, expiry) emits a `StockChanged` event.
4. A reservation expires 15 minutes after it is made (`RESERVATION_TTL_MS`); an expired reservation no longer counts as reserved.
5. The stock slice never imports from the checkout slice.

## Deferred / Open questions

- Committing a reservation that has already expired does nothing; whether it should fail is undecided.

## Acceptance criteria

- `tests/stock.test.ts` passes.

---

Last updated: 2026-10-25.
