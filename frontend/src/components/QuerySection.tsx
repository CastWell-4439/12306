import { FormEvent } from "react";
import { ExecuteFn } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  execute: ExecuteFn;
  orderId: string;
  queryOrderId: string;
  setQueryOrderId: (v: string) => void;
}

export function QuerySection({ execute, orderId, queryOrderId, setQueryOrderId }: Props) {
  return (
    <SectionCard title="查询模块" desc="查询 query-service 维护的订单读模型。">
      <form
        className="form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault();
          void execute({
            label: "Query Order View",
            service: "query",
            method: "GET",
            path: `/query/orders?order_id=${encodeURIComponent(queryOrderId)}`
          });
        }}
      >
        <label>
          order_id
          <input value={queryOrderId} onChange={(e) => setQueryOrderId(e.target.value)} />
        </label>
        <div className="inline-actions">
          <button type="button" className="btn-secondary" onClick={() => setQueryOrderId(orderId)}>
            使用当前 order_id
          </button>
          <button type="submit">查询读模型</button>
        </div>
      </form>
    </SectionCard>
  );
}


