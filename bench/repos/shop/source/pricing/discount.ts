import { type Money, money, min, subtract } from "./money.ts";

// Half up: add half of the divisor before the integer division.
export function percentOff(amount: Money, percent: number): Money {
  if (!Number.isInteger(percent) || percent < 0 || percent > 100) {
    throw new RangeError("percent must be an integer from 0 to 100");
  }
  return money(Math.floor((amount.cents * percent + 50) / 100), amount.currency);
}

export function applyDiscount(subtotal: Money, discount: Money): Money {
  return subtract(subtotal, min(discount, subtotal));
}
