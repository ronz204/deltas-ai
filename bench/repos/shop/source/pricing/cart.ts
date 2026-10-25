import { type Money, add, money } from "./money.ts";
import { applyDiscount, percentOff } from "./discount.ts";
import { taxOn } from "./tax.ts";

export type Line = { sku: string; qty: number; unit: Money };

export type Pricing = {
  subtotal: Money;
  discount: Money;
  taxable: Money;
  tax: Money;
  total: Money;
};

export function lineTotal(line: Line): Money {
  if (!Number.isInteger(line.qty) || line.qty < 1) {
    throw new RangeError("qty must be a positive integer");
  }
  if (line.unit.cents < 0) throw new RangeError("unit price must not be negative");
  return money(line.unit.cents * line.qty, line.unit.currency);
}

export function priceCart(
  lines: Line[],
  opts: { percentOff?: number; taxBps: number },
): Pricing {
  let subtotal = money(0, lines[0]?.unit.currency ?? "USD");
  for (const line of lines) subtotal = add(subtotal, lineTotal(line));
  const discount = percentOff(subtotal, opts.percentOff ?? 0);
  const taxable = applyDiscount(subtotal, discount);
  const tax = taxOn(taxable, opts.taxBps);
  return { subtotal, discount, taxable, tax, total: add(taxable, tax) };
}
