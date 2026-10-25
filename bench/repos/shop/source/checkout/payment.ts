import type { Money } from "@pricing";

export type ChargeResult = { ok: true } | { ok: false; reason: string };

export type PaymentGateway = {
  charge(orderId: string, amount: Money): Promise<ChargeResult>;
};
