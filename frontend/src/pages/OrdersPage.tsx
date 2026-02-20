import { OrderSection } from "../components/OrderSection";
import { QuerySection } from "../components/QuerySection";
import { ResultPanel } from "../components/ResultPanel";
import { useExecutor } from "../hooks/useExecutor";
import { useAppState } from "../state/AppState";

export function OrdersPage() {
  const { execute } = useExecutor();
  const { drafts, setDrafts, result } = useAppState();

  return (
    <>
      <div className="two-col">
        <OrderSection
          execute={execute}
          idempotencyKey={drafts.idempotencyKey}
          setIdempotencyKey={(v) => setDrafts((d) => ({ ...d, idempotencyKey: v }))}
          amountCents={drafts.amountCents}
          setAmountCents={(v) => setDrafts((d) => ({ ...d, amountCents: v }))}
          orderId={drafts.orderId}
          setOrderId={(v) => setDrafts((d) => ({ ...d, orderId: v }))}
          providerTxnId={drafts.providerTxnId}
          setProviderTxnId={(v) => setDrafts((d) => ({ ...d, providerTxnId: v }))}
          paymentStatus={drafts.paymentStatus}
          setPaymentStatus={(v) => setDrafts((d) => ({ ...d, paymentStatus: v }))}
        />
        <QuerySection
          execute={execute}
          orderId={drafts.orderId}
          queryOrderId={drafts.queryOrderId}
          setQueryOrderId={(v) => setDrafts((d) => ({ ...d, queryOrderId: v }))}
        />
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


