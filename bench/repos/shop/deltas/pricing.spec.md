# Pricing — Spec

> Turns cart lines into a priced breakdown in integer cents: subtotal, discount, tax, total.

## Intent

Every amount a customer sees must be exact and reproducible. Pricing is the single place where money is computed, so the rest of the system can treat its results as final.

## Scope

Owns money values, percentage discounts, tax, and cart pricing. Does not own stock, payment, or order persistence. Does not own currency conversion: amounts in different currencies are never combined.

## Contract

| Function | Behavior |
|---|---|
| `usd(cents)` / `money(cents, currency)` | Builds a `Money`; `cents` must be an integer. |
| `percentOff(amount, percent)` | `percent` is an integer from 0 to 100; returns the discount amount. |
| `applyDiscount(subtotal, discount)` | Returns the subtotal minus the discount, with the discount capped at the subtotal. |
| `taxOn(amount, rateBps)` | `rateBps` is a non-negative integer in basis points (1000 = 10%). |
| `priceCart(lines, { percentOff?, taxBps })` | Returns `{ subtotal, discount, taxable, tax, total }`. A line has a positive integer quantity and a non-negative unit price. |

## Invariants

1. Money is integer cents: price arithmetic never uses floating-point literals, `toFixed`, or `parseFloat`.
2. Percentage discounts round half up to the nearest cent.
3. A total is never negative: a discount is capped at the subtotal.
4. Tax is computed on the discounted subtotal, not on the original subtotal.
5. The pricing slice is pure: it performs no I/O, reads no clock, and uses no randomness.
6. Combining amounts in different currencies throws `CurrencyMismatchError`.

## Deferred / Open questions

- Multiple tax rates per cart are not supported; a cart has one rate.

## Acceptance criteria

- `tests/pricing.test.ts` passes.

---

Last updated: 2026-10-25.
