export type Money = { readonly cents: number; readonly currency: string };

export class CurrencyMismatchError extends Error {
  constructor(a: string, b: string) {
    super(`currency mismatch: ${a} vs ${b}`);
    this.name = "CurrencyMismatchError";
  }
}

export function money(cents: number, currency: string): Money {
  if (!Number.isInteger(cents)) throw new RangeError("cents must be an integer");
  return { cents, currency };
}

export function usd(cents: number): Money {
  return money(cents, "USD");
}

function assertSameCurrency(a: Money, b: Money): void {
  if (a.currency !== b.currency) throw new CurrencyMismatchError(a.currency, b.currency);
}

export function add(a: Money, b: Money): Money {
  assertSameCurrency(a, b);
  return money(a.cents + b.cents, a.currency);
}

export function subtract(a: Money, b: Money): Money {
  assertSameCurrency(a, b);
  return money(a.cents - b.cents, a.currency);
}

export function min(a: Money, b: Money): Money {
  assertSameCurrency(a, b);
  return a.cents <= b.cents ? a : b;
}
