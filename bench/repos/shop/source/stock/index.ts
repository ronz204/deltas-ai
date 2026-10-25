export type { EventSink, StockChanged } from "./events.ts";
export {
  type Inventory,
  InsufficientStockError,
  RESERVATION_TTL_MS,
  createInventory,
} from "./inventory.ts";
