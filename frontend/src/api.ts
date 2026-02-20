import { API_PREFIX, ApiCallInput, ApiCallResult, JsonValue } from "./types";

export async function request(input: ApiCallInput): Promise<ApiCallResult> {
  const start = performance.now();
  const response = await fetch(`${API_PREFIX[input.service]}${input.path}`, {
    method: input.method,
    headers: {
      "Content-Type": "application/json"
    },
    body: input.body === undefined ? undefined : JSON.stringify(input.body)
  });
  const durationMs = Math.round(performance.now() - start);
  const text = await response.text();

  let data: JsonValue = text;
  try {
    data = text ? (JSON.parse(text) as JsonValue) : null;
  } catch {
    // keep raw text fallback
  }

  return {
    ok: response.ok,
    status: response.status,
    durationMs,
    data
  };
}
