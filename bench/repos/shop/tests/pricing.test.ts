import { describe, expect, test } from "bun:test";
import {
  CurrencyMismatchError,
  add,
  applyDiscount,
  money,
  percentOff,
  priceCart,
  usd,
} from "../src/pricing/index.ts";

describe("percentOff", () => {
  test("rounds half up to the nearest cent", () => {
    expect(percentOff(usd(105), 10).cents).toBe(11);
    expect(percentOff(usd(50), 1).cents).toBe(1);
    expect(percentOff(usd(104), 10).cents).toBe(10);
  });

  test("is exact when there is no fraction", () => {
    expect(percentOff(usd(250), 10).cents).toBe(25);
  });
});

describe("applyDiscount", () => {
  test("never goes below zero", () => {
    expect(applyDiscount(usd(500), usd(900)).cents).toBe(0);
  });

  test("subtracts a smaller discount", () => {
    expect(applyDiscount(usd(500), usd(120)).cents).toBe(380);
  });
});

describe("priceCart", () => {
  const lines = [{ sku: "a", qty: 2, unit: usd(5000) }];

  test("computes tax on the discounted subtotal", () => {
    const priced = priceCart(lines, { percentOff: 10, taxBps: 1000 });
    expect(priced.subtotal.cents).toBe(10000);
    expect(priced.taxable.cents).toBe(9000);
    expect(priced.tax.cents).toBe(900);
    expect(priced.total.cents).toBe(9900);
  });

  test("a negative unit price is rejected", () => {
    const bad = [{ sku: "a", qty: 1, unit: usd(-100) }];
    expect(() => priceCart(bad, { taxBps: 0 })).toThrow(RangeError);
  });

  test("a full discount leaves a zero total", () => {
    const priced = priceCart(lines, { percentOff: 100, taxBps: 1000 });
    expect(priced.total.cents).toBe(0);
  });
});

describe("currencies", () => {
  test("adding different currencies throws", () => {
    expect(() => add(usd(100), money(100, "EUR"))).toThrow(CurrencyMismatchError);
  });

  test("a cart mixing currencies throws", () => {
    const lines = [
      { sku: "a", qty: 1, unit: usd(100) },
      { sku: "b", qty: 1, unit: money(100, "EUR") },
    ];
    expect(() => priceCart(lines, { taxBps: 0 })).toThrow(CurrencyMismatchError);
  });
});
