import { FormEvent } from "react";
import { ExecuteFn } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  execute: ExecuteFn;
  partitionKey: string;
  setPartitionKey: (v: string) => void;
  holdId: string;
  setHoldId: (v: string) => void;
  qty: number;
  setQty: (v: number) => void;
  capacity: number;
  setCapacity: (v: number) => void;
}

export function InventorySection(props: Props) {
  return (
    <SectionCard title="库存模块" desc="围绕 hold 进行尝试占座、释放、确认与可用量查询。">
      <form
        className="form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault();
          void props.execute({
            label: "Try Hold",
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
        }}
      >
        <h3>1) Try Hold</h3>
        <label>
          partition_key
          <input value={props.partitionKey} onChange={(e) => props.setPartitionKey(e.target.value)} />
        </label>
        <label>
          hold_id
          <input value={props.holdId} onChange={(e) => props.setHoldId(e.target.value)} />
        </label>
        <label>
          qty
          <input type="number" value={props.qty} onChange={(e) => props.setQty(Number(e.target.value))} />
        </label>
        <label>
          capacity
          <input
            type="number"
            value={props.capacity}
            onChange={(e) => props.setCapacity(Number(e.target.value))}
          />
        </label>
        <button type="submit">尝试占座</button>
      </form>

      <div className="inline-actions">
        <button
          onClick={() =>
            props.execute({
              label: "Release Hold",
              service: "inventory",
              method: "POST",
              path: "/inventory/release-hold",
              body: {
                partition_key: props.partitionKey,
                hold_id: props.holdId
              }
            })
          }
        >
          释放 hold
        </button>
        <button
          onClick={() =>
            props.execute({
              label: "Confirm Hold",
              service: "inventory",
              method: "POST",
              path: "/inventory/confirm-hold",
              body: {
                partition_key: props.partitionKey,
                hold_id: props.holdId
              }
            })
          }
        >
          确认 hold
        </button>
        <button
          onClick={() =>
            props.execute({
              label: "Get Availability",
              service: "inventory",
              method: "GET",
              path: `/inventory/availability?partition_key=${encodeURIComponent(props.partitionKey)}`
            })
          }
        >
          查询可用量
        </button>
      </div>
    </SectionCard>
  );
}


