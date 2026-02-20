export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [k: string]: JsonValue };

export type ServiceKey =
  | "gateway"
  | "order"
  | "inventory"
  | "query"
  | "worker"
  | "nginx";

export const API_PREFIX: Record<ServiceKey, string> = {
  gateway: "/api/gateway",
  order: "/api/order",
  inventory: "/api/inventory",
  query: "/api/query",
  worker: "/api/worker",
  nginx: "/api/nginx"
};

export interface ApiCallInput {
  service: ServiceKey;
  method: "GET" | "POST";
  path: string;
  body?: JsonValue;
}

export interface ApiCallResult {
  ok: boolean;
  status: number;
  durationMs: number;
  data: JsonValue;
}

export interface RequestHistoryItem {
  id: string;
  at: string;
  label: string;
  service: ServiceKey;
  method: "GET" | "POST";
  path: string;
  body?: JsonValue;
  ok: boolean;
  status: number;
  durationMs: number;
}

export interface ExecuteActionInput {
  label: string;
  service: ServiceKey;
  method: "GET" | "POST";
  path: string;
  body?: JsonValue;
  onSuccess?: (data: JsonValue) => void;
}

export type ExecuteFn = (input: ExecuteActionInput) => Promise<JsonValue>;

export interface SessionState {
  username: string;
  token: string;
}
