import type { EventSink } from "./events.ts";

export const RESERVATION_TTL_MS = 15 * 60 * 1000;

export class InsufficientStockError extends Error {
  constructor(sku: string) {
    super(`insufficient stock for ${sku}`);
    this.name = "InsufficientStockError";
  }
}

type Reservation = { orderId: string; sku: string; qty: number; expiresAt: number };

export type Inventory = ReturnType<typeof createInventory>;

export function createInventory(deps: { now: () => number; emit: EventSink }) {
  const onHand = new Map<string, number>();
  const reservations = new Map<string, Reservation>();

  function reservedFor(sku: string): number {
    let total = 0;
    for (const r of reservations.values()) if (r.sku === sku) total += r.qty;
    return total;
  }

  function changed(sku: string): void {
    deps.emit({
      type: "StockChanged",
      sku,
      onHand: onHand.get(sku) ?? 0,
      reserved: reservedFor(sku),
      at: deps.now(),
    });
  }

  function expire(): void {
    const t = deps.now();
    for (const [key, r] of reservations) {
      if (r.expiresAt <= t) {
        reservations.delete(key);
        changed(r.sku);
      }
    }
  }

  return {
    receive(sku: string, qty: number): void {
      if (!Number.isInteger(qty) || qty < 1) throw new RangeError("qty must be a positive integer");
      onHand.set(sku, (onHand.get(sku) ?? 0) + qty);
      changed(sku);
    },

    available(sku: string): number {
      expire();
      return (onHand.get(sku) ?? 0) - reservedFor(sku);
    },

    reserve(orderId: string, sku: string, qty: number): void {
      expire();
      const key = `${orderId}:${sku}`;
      if (reservations.has(key)) return;
      if ((onHand.get(sku) ?? 0) - reservedFor(sku) < qty) throw new InsufficientStockError(sku);
      reservations.set(key, { orderId, sku, qty, expiresAt: deps.now() + RESERVATION_TTL_MS });
      changed(sku);
    },

    release(orderId: string): void {
      for (const [key, r] of reservations) {
        if (r.orderId === orderId) {
          reservations.delete(key);
          changed(r.sku);
        }
      }
    },

    commit(orderId: string): void {
      for (const [key, r] of reservations) {
        if (r.orderId === orderId) {
          onHand.set(r.sku, (onHand.get(r.sku) ?? 0) - r.qty);
          reservations.delete(key);
          changed(r.sku);
        }
      }
    },
  };
}
