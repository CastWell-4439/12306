import { JsonValue } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  title: string;
  data: JsonValue;
  status?: number;
  durationMs?: number;
}

export function ResultPanel({ title, data, status, durationMs }: Props) {
  return (
    <SectionCard title="结果面板" desc="显示最近一次请求的完整响应。">
      <div className="result-meta">
        <strong>{title}</strong>
        {status !== undefined ? <span>HTTP {status}</span> : null}
        {durationMs !== undefined ? <span>{durationMs}ms</span> : null}
      </div>
      <pre>{JSON.stringify(data, null, 2)}</pre>
    </SectionCard>
  );
}


