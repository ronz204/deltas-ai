import { type Line, priceCart } from "@pricing";
import type { Inventory } from "@stock";
import type { PaymentGateway } from "./payment.ts";
import type { Order, OrderStore } from "./orders.ts";

export class EmptyCartError extends Error {
  constructor() {
    super("cart is empty");
    this.name = "EmptyCartError";
  }
}

export class DuplicateOrderError extends Error {
  constructor(orderId: string) {
    super(`order already attempted: ${orderId}`);
    this.name = "DuplicateOrderError";
  }
}

export class PaymentFailedError extends Error {
  constructor(reason: string) {
    super(`payment failed: ${reason}`);
    this.name = "PaymentFailedError";
  }
}

export type CheckoutRequest = {
  orderId: string;
  lines: Line[];
  percentOff?: number;
  taxBps: number;
};

export function createCheckout(deps: {
  stock: Inventory;
  gateway: PaymentGateway;
  orders: OrderStore;
}) {
  const attempted = new Set<string>();

  return async function checkout(req: CheckoutRequest): Promise<Order> {
    if (req.lines.length === 0) throw new EmptyCartError();
    if (attempted.has(req.orderId)) throw new DuplicateOrderError(req.orderId);

    const pricing = priceCart(req.lines, { percentOff: req.percentOff, taxBps: req.taxBps });

    try {
      for (const line of req.lines) deps.stock.reserve(req.orderId, line.sku, line.qty);
    } catch (error) {
      deps.stock.release(req.orderId);
      throw error;
    }

    attempted.add(req.orderId);
    const result = await deps.gateway.charge(req.orderId, pricing.total);
    if (!result.ok) {
      deps.stock.release(req.orderId);
      throw new PaymentFailedError(result.reason);
    }

    deps.stock.commit(req.orderId);
    const order: Order = { orderId: req.orderId, lines: req.lines, pricing };
    await deps.orders.save(order);
    return order;
  };
}
