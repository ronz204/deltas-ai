import { describe, expect, test } from "bun:test";
import { InsufficientStockError, createInventory } from "@stock";

function setup() {
  const stock = createInventory({ now: () => 0, emit: () => {} });
  stock.receive("sku-1", 10);
  return stock;
}

describe("reservations", () => {
  test("reserving twice for the same order counts once", () => {
    const stock = setup();
    stock.reserve("o1", "sku-1", 3);
    stock.reserve("o1", "sku-1", 3);
    expect(stock.available("sku-1")).toBe(7);
  });

  test("different orders each reserve their own quantity", () => {
    const stock = setup();
    stock.reserve("o1", "sku-1", 3);
    stock.reserve("o2", "sku-1", 3);
    expect(stock.available("sku-1")).toBe(4);
  });
});

describe("availability", () => {
  test("available is on hand minus reserved", () => {
    const stock = setup();
    stock.reserve("o1", "sku-1", 4);
    expect(stock.available("sku-1")).toBe(6);
  });

  test("a reservation beyond what is available is rejected and never goes negative", () => {
    const stock = setup();
    expect(() => stock.reserve("o1", "sku-1", 11)).toThrow(InsufficientStockError);
    expect(stock.available("sku-1")).toBe(10);
  });

  test("releasing gives the quantity back", () => {
    const stock = setup();
    stock.reserve("o1", "sku-1", 4);
    stock.release("o1");
    expect(stock.available("sku-1")).toBe(10);
  });

  test("committing consumes the quantity for good", () => {
    const stock = setup();
    stock.reserve("o1", "sku-1", 4);
    stock.commit("o1");
    expect(stock.available("sku-1")).toBe(6);
  });
});
