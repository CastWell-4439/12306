import { RequestHistoryItem } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  items: RequestHistoryItem[];
  onClear: () => void;
  onReplay: (item: RequestHistoryItem) => void;
}

export function HistoryPanel({ items, onClear, onReplay }: Props) {
  return (
    <SectionCard title="请求历史" desc="用于排查接口调用顺序和耗时。">
      <div className="toolbar">
        <button onClick={onClear} className="btn-secondary">
          清空历史
        </button>
      </div>
      <div className="history-list">
        {items.length === 0 ? (
          <p className="empty">暂无请求记录</p>
        ) : (
          items.map((item) => (
            <div key={item.id} className={`history-item ${item.ok ? "ok" : "fail"}`}>
              <div>
                <strong>{item.label}</strong>
                <div className="sub">
                  {item.method} {item.path} ({item.service})
                </div>
              </div>
              <div className="history-meta">
                <span>HTTP {item.status}</span>
                <span>{item.durationMs}ms</span>
                <span>{item.at}</span>
                <button className="btn-secondary" onClick={() => onReplay(item)}>
                  重放
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </SectionCard>
  );
}


