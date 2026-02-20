import { ExecuteFn } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  execute: ExecuteFn;
  orderId: string;
  setOrderId: (v: string) => void;
  idempotencyKey: string;
  amountCents: number;
  providerTxnId: string;
  paymentStatus: string;
  partitionKey: string;
  holdId: string;
  qty: number;
  capacity: number;
}

function extractOrderId(data: unknown): string | null {
  if (data && typeof data === "object" && "OrderID" in data) {
    return String((data as { OrderID: unknown }).OrderID);
  }
  return null;
}

export function WorkflowSection(props: Props) {
  return (
    <SectionCard
      title="流程编排"
      desc="一键执行典型链路，减少手工逐步点击：下单 -> 预占 -> 支付 -> 查读模型。"
    >
      <div className="inline-actions">
        <button
          onClick={async () => {
            const flowKey = `${props.idempotencyKey}-${Date.now()}`;
            const txnId = `${props.providerTxnId}-${Date.now()}`;

            const create = await props.execute({
              label: "Flow/Create Order",
              service: "order",
              method: "POST",
              path: "/orders",
              body: {
                idempotency_key: flowKey,
                amount_cents: props.amountCents
              }
            });
            const createdOrderID = extractOrderId(create);
            if (!createdOrderID) return;
            props.setOrderId(createdOrderID);

            await props.execute({
              label: "Flow/Reserve Order",
              service: "order",
              method: "POST",
              path: "/orders/reserve",
              body: { order_id: createdOrderID }
            });

            await props.execute({
              label: "Flow/Payment Callback",
              service: "order",
              method: "POST",
              path: "/payments/callback",
              body: {
                order_id: createdOrderID,
                provider_txn_id: txnId,
                status: props.paymentStatus
              }
            });

            await props.execute({
              label: "Flow/Query Order View",
              service: "query",
              method: "GET",
              path: `/query/orders?order_id=${encodeURIComponent(createdOrderID)}`
            });
          }}
        >
          执行订单主流程
        </button>

        <button
          onClick={async () => {
            await props.execute({
              label: "Flow/Try Hold",
              service: "inventory",
              method: "POST",
              path: "/inventory/try-hold",
              body: {
                partition_key: props.partitionKey,
                hold_id: props.holdId,
                qty: props.qty,
                capacity: props.capacity
              }
            });
            await props.execute({
              label: "Flow/Confirm Hold",
              service: "inventory",
              method: "POST",
              path: "/inventory/confirm-hold",
              body: {
                partition_key: props.partitionKey,
                hold_id: props.holdId
              }
            });
            await props.execute({
              label: "Flow/Get Availability",
              service: "inventory",
              method: "GET",
              path: `/inventory/availability?partition_key=${encodeURIComponent(props.partitionKey)}`
            });
          }}
        >
          执行库存流程
        </button>
      </div>
    </SectionCard>
  );
}


