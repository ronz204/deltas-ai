import type { Line, Pricing } from "../pricing/index.ts";

export type Order = { orderId: string; lines: Line[]; pricing: Pricing };

export type OrderStore = {
  save(order: Order): Promise<void>;
};
