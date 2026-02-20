import { request } from "../api";
import { ExecuteActionInput, JsonValue, RequestHistoryItem } from "../types";
import { useAppState } from "../state/AppState";

export function useExecutor() {
  const { setResult, setHistory } = useAppState();

  const execute = async (input: ExecuteActionInput): Promise<JsonValue> => {
    const res = await request({
      service: input.service,
      method: input.method,
      path: input.path,
      body: input.body
    });

    const historyItem: RequestHistoryItem = {
      id: `${Date.now()}-${Math.random()}`,
      at: new Date().toLocaleTimeString(),
      label: input.label,
      service: input.service,
      method: input.method,
      path: input.path,
      body: input.body,
      ok: res.ok,
      status: res.status,
      durationMs: res.durationMs
    };

    setResult({
      title: input.label,
      data: res.data,
      status: res.status,
      durationMs: res.durationMs
    });
    setHistory((h) => [historyItem, ...h].slice(0, 100));

    if (!res.ok) {
      throw new Error(
        `HTTP ${res.status}: ${
          typeof res.data === "string" ? res.data : JSON.stringify(res.data)
        }`
      );
    }
    if (input.onSuccess) input.onSuccess(res.data);
    return res.data;
  };

  const replay = async (historyItem: RequestHistoryItem): Promise<JsonValue> => {
    return execute({
      label: `${historyItem.label} (replay)`,
      service: historyItem.service,
      method: historyItem.method,
      path: historyItem.path,
      body: historyItem.body
    });
  };

  return { execute, replay };
}


