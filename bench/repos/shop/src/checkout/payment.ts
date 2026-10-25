import type { Money } from "../pricing/index.ts";

export type ChargeResult = { ok: true } | { ok: false; reason: string };

export type PaymentGateway = {
  charge(orderId: string, amount: Money): Promise<ChargeResult>;
};
