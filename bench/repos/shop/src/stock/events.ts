export type StockChanged = {
  type: "StockChanged";
  sku: string;
  onHand: number;
  reserved: number;
  at: number;
};

export type EventSink = (event: StockChanged) => void;
