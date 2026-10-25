# Shop

A small TypeScript backend for a shop: pricing, stock, and checkout. Runs on Bun; no dependencies.

## Layout

- `source/pricing/` — money, discounts, tax, cart pricing.
- `source/stock/` — on-hand quantities, reservations, stock events.
- `source/checkout/` — orchestrates pricing, stock, and payment into an order.
- `tests/` — `bun test`.

## Contracts

Each slice has a living contract under `deltas/<slice>.spec.md`. Read the matching spec before changing a slice, and verify it still holds afterwards.

## Commands

- `bun test` — run the whole suite.
