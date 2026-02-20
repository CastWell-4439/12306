import { useEffect, useState } from "react";
import { QuerySection } from "../components/QuerySection";
import { ResultPanel } from "../components/ResultPanel";
import { WorkflowSection } from "../components/WorkflowSection";
import { useExecutor } from "../hooks/useExecutor";
import { useAppState } from "../state/AppState";

export function BookingPage() {
  const { execute } = useExecutor();
  const { drafts, setDrafts, result } = useAppState();

  const [polling, setPolling] = useState(false);

  useEffect(() => {
    if (!polling || !drafts.orderId) return;
    const id = setInterval(() => {
      void execute({
        label: "Poll Query Order View",
        service: "query",
        method: "GET",
        path: `/query/orders?order_id=${encodeURIComponent(drafts.orderId)}`
      });
    }, 3000);
    return () => clearInterval(id);
  }, [polling, drafts.orderId, execute]);

  return (
    <>
      <WorkflowSection
        execute={execute}
        orderId={drafts.orderId}
        setOrderId={(v) => setDrafts((d) => ({ ...d, orderId: v }))}
        idempotencyKey={drafts.idempotencyKey}
        amountCents={drafts.amountCents}
        providerTxnId={drafts.providerTxnId}
        paymentStatus={drafts.paymentStatus}
        partitionKey={drafts.partitionKey}
        holdId={drafts.holdId}
        qty={drafts.qty}
        capacity={drafts.capacity}
      />

      <QuerySection
        execute={execute}
        orderId={drafts.orderId}
        queryOrderId={drafts.queryOrderId}
        setQueryOrderId={(v) => setDrafts((d) => ({ ...d, queryOrderId: v }))}
      />

      <div className="card">
        <h2>订单状态轮询</h2>
        <p>用于模拟支付后持续查询状态，便于观察 Ticketed 等状态变化。</p>
        <div className="inline-actions">
          <button
            className={polling ? "btn-secondary" : ""}
            onClick={() => setPolling((v) => !v)}
          >
            {polling ? "停止轮询" : "开始轮询(3s)"}
          </button>
        </div>
      </div>

      <ResultPanel
        title={result.title}
        data={result.data}
        status={result.status}
        durationMs={result.durationMs}
      />
    </>
  );
}


