import { FormEvent } from "react";
import { ExecuteFn } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  execute: ExecuteFn;
  idempotencyKey: string;
  setIdempotencyKey: (v: string) => void;
  amountCents: number;
  setAmountCents: (v: number) => void;
  orderId: string;
  setOrderId: (v: string) => void;
  providerTxnId: string;
  setProviderTxnId: (v: string) => void;
  paymentStatus: string;
  setPaymentStatus: (v: string) => void;
}

export function OrderSection(props: Props) {
  return (
    <SectionCard title="订单模块" desc="创建订单、预占、支付回调、查询订单明细。">
      <form
        className="form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault();
          void props.execute({
            label: "Create Order",
            service: "order",
            method: "POST",
            path: "/orders",
            body: {
              idempotency_key: props.idempotencyKey,
              amount_cents: props.amountCents
            },
            onSuccess: (data) => {
              if (data && typeof data === "object" && "OrderID" in data) {
                props.setOrderId(String(data.OrderID));
              }
            }
          });
        }}
      >
        <h3>1) 创建订单</h3>
        <label>
          idempotency_key
          <input
            value={props.idempotencyKey}
            onChange={(e) => props.setIdempotencyKey(e.target.value)}
          />
        </label>
        <label>
          amount_cents
          <input
            type="number"
            value={props.amountCents}
            onChange={(e) => props.setAmountCents(Number(e.target.value))}
          />
        </label>
        <button type="submit">创建订单</button>
      </form>

      <form
        className="form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault();
          void props.execute({
            label: "Reserve Order",
            service: "order",
            method: "POST",
            path: "/orders/reserve",
            body: { order_id: props.orderId }
          });
        }}
      >
        <h3>2) 预占订单</h3>
        <label>
          order_id
          <input value={props.orderId} onChange={(e) => props.setOrderId(e.target.value)} />
        </label>
        <button type="submit">预占</button>
      </form>

      <form
        className="form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault();
          void props.execute({
            label: "Payment Callback",
            service: "order",
            method: "POST",
            path: "/payments/callback",
            body: {
              order_id: props.orderId,
              provider_txn_id: props.providerTxnId,
              status: props.paymentStatus
            }
          });
        }}
      >
        <h3>3) 支付回调</h3>
        <label>
          provider_txn_id
          <input
            value={props.providerTxnId}
            onChange={(e) => props.setProviderTxnId(e.target.value)}
          />
        </label>
        <label>
          status
          <input value={props.paymentStatus} onChange={(e) => props.setPaymentStatus(e.target.value)} />
        </label>
        <button type="submit">支付回调</button>
      </form>

      <div className="inline-actions">
        <button
          onClick={() =>
            props.execute({
              label: "Get Order",
              service: "order",
              method: "GET",
              path: `/orders/get?order_id=${encodeURIComponent(props.orderId)}`
            })
          }
        >
          查询订单详情
        </button>
      </div>
    </SectionCard>
  );
}


