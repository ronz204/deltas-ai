import { type Money, money } from "./money.ts";

// The rate is in basis points (1000 = 10%). Half up, in integer cents.
export function taxOn(amount: Money, rateBps: number): Money {
  if (!Number.isInteger(rateBps) || rateBps < 0) {
    throw new RangeError("rateBps must be a non-negative integer");
  }
  return money(Math.floor((amount.cents * rateBps + 5000) / 10000), amount.currency);
}
