import { InventorySection } from "../components/InventorySection";
import { ResultPanel } from "../components/ResultPanel";
import { useExecutor } from "../hooks/useExecutor";
import { useAppState } from "../state/AppState";

export function InventoryPage() {
  const { execute } = useExecutor();
  const { drafts, setDrafts, result } = useAppState();

  return (
    <>
      <InventorySection
        execute={execute}
        partitionKey={drafts.partitionKey}
        setPartitionKey={(v) => setDrafts((d) => ({ ...d, partitionKey: v }))}
        holdId={drafts.holdId}
        setHoldId={(v) => setDrafts((d) => ({ ...d, holdId: v }))}
        qty={drafts.qty}
        setQty={(v) => setDrafts((d) => ({ ...d, qty: v }))}
        capacity={drafts.capacity}
        setCapacity={(v) => setDrafts((d) => ({ ...d, capacity: v }))}
      />
      <ResultPanel
        title={result.title}
        data={result.data}
        status={result.status}
        durationMs={result.durationMs}
      />
    </>
  );
}


