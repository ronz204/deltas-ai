import type { Line, Pricing } from "@pricing";

export type Order = { orderId: string; lines: Line[]; pricing: Pricing };

export type OrderStore = {
  save(order: Order): Promise<void>;
};
