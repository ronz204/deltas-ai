import { describe, expect, test } from "bun:test";
import {
  DuplicateOrderError,
  EmptyCartError,
  PaymentFailedError,
  createCheckout,
  type ChargeResult,
  type Order,
} from "@checkout";
import { usd } from "@pricing";
import { createInventory } from "@stock";

function setup(result: ChargeResult = { ok: true }) {
  const stock = createInventory({ now: () => 0, emit: () => {} });
  stock.receive("sku-1", 10);

  const saved: Order[] = [];
  const charges: { orderId: string; available: number; savedAtCharge: number }[] = [];

  const gateway = {
    async charge(orderId: string) {
      charges.push({ orderId, available: stock.available("sku-1"), savedAtCharge: saved.length });
      return result;
    },
  };
  const orders = {
    async save(order: Order) {
      saved.push(order);
    },
  };

  const checkout = createCheckout({ stock, gateway, orders });
  const request = (orderId = "o1") => ({
    orderId,
    lines: [{ sku: "sku-1", qty: 2, unit: usd(1000) }],
    taxBps: 1000,
  });
  return { stock, saved, charges, checkout, request };
}

describe("checkout", () => {
  test("an empty cart throws EmptyCartError", async () => {
    const { checkout } = setup();
    await expect(checkout({ orderId: "o1", lines: [], taxBps: 0 })).rejects.toThrow(
      EmptyCartError,
    );
  });

  test("reserves stock before charging", async () => {
    const { charges, checkout, request } = setup();
    await checkout(request());
    expect(charges).toHaveLength(1);
    expect(charges[0]!.available).toBe(8);
  });

  test("releases the reservation when payment fails", async () => {
    const { stock, checkout, request } = setup({ ok: false, reason: "declined" });
    await expect(checkout(request())).rejects.toThrow(PaymentFailedError);
    expect(stock.available("sku-1")).toBe(10);
  });

  test("attempts payment once even when it fails", async () => {
    const { charges, checkout, request } = setup({ ok: false, reason: "declined" });
    await expect(checkout(request())).rejects.toThrow(PaymentFailedError);
    expect(charges).toHaveLength(1);
  });

  test("a second checkout for the same order is rejected without charging again", async () => {
    const { charges, checkout, request } = setup();
    await checkout(request());
    await expect(checkout(request())).rejects.toThrow(DuplicateOrderError);
    expect(charges).toHaveLength(1);
  });

  test("persists the order only after payment succeeds", async () => {
    const { saved, charges, checkout, request } = setup();
    await checkout(request());
    expect(charges[0]!.savedAtCharge).toBe(0);
    expect(saved).toHaveLength(1);
  });

  test("persists nothing when payment fails", async () => {
    const { saved, checkout, request } = setup({ ok: false, reason: "declined" });
    await expect(checkout(request())).rejects.toThrow(PaymentFailedError);
    expect(saved).toHaveLength(0);
  });
});
