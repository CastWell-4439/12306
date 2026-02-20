import { HealthSection } from "../components/HealthSection";
import { HistoryPanel } from "../components/HistoryPanel";
import { ResultPanel } from "../components/ResultPanel";
import { useExecutor } from "../hooks/useExecutor";
import { useAppState } from "../state/AppState";

export function DashboardPage() {
  const { execute, replay } = useExecutor();
  const { result, history, setHistory } = useAppState();

  return (
    <>
      <HealthSection execute={execute} />
      <div className="two-col">
        <ResultPanel
          title={result.title}
          data={result.data}
          status={result.status}
          durationMs={result.durationMs}
        />
        <HistoryPanel
          items={history}
          onClear={() => setHistory([])}
          onReplay={(item) => void replay(item)}
        />
      </div>
    </>
  );
}


