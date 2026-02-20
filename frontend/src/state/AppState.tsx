import {
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState
} from "react";
import { JsonValue, RequestHistoryItem, SessionState } from "../types";

interface ResultState {
  title: string;
  data: JsonValue;
  status?: number;
  durationMs?: number;
}

interface DraftState {
  orderId: string;
  idempotencyKey: string;
  amountCents: number;
  providerTxnId: string;
  paymentStatus: string;
  queryOrderId: string;
  partitionKey: string;
  holdId: string;
  qty: number;
  capacity: number;
}

interface AppStateContextValue {
  session: SessionState | null;
  setSession: Dispatch<SetStateAction<SessionState | null>>;
  result: ResultState;
  setResult: Dispatch<SetStateAction<ResultState>>;
  history: RequestHistoryItem[];
  setHistory: Dispatch<SetStateAction<RequestHistoryItem[]>>;
  drafts: DraftState;
  setDrafts: Dispatch<SetStateAction<DraftState>>;
}

const SESSION_KEY = "ticketing.frontend.session";
const HISTORY_KEY = "ticketing.frontend.history";
const DRAFTS_KEY = "ticketing.frontend.drafts";

const AppStateContext = createContext<AppStateContextValue | null>(null);

const initialDrafts: DraftState = {
  idempotencyKey: "demo-key-1",
  amountCents: 10000,
  orderId: "",
  providerTxnId: "txn-demo-1",
  paymentStatus: "SUCCESS",
  queryOrderId: "",
  partitionKey: "G123|2026-02-11|2nd",
  holdId: "hold-demo-1",
  qty: 1,
  capacity: 200
};

function safeParse<T>(raw: string | null, fallback: T): T {
  if (!raw) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export function AppStateProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<SessionState | null>(() =>
    safeParse<SessionState | null>(localStorage.getItem(SESSION_KEY), null)
  );
  const [history, setHistory] = useState<RequestHistoryItem[]>(() =>
    safeParse<RequestHistoryItem[]>(localStorage.getItem(HISTORY_KEY), [])
  );
  const [drafts, setDrafts] = useState<DraftState>(() =>
    safeParse<DraftState>(localStorage.getItem(DRAFTS_KEY), initialDrafts)
  );
  const [result, setResult] = useState<ResultState>({
    title: "Ready",
    data: { message: "Welcome. Please execute an action." }
  });

  useEffect(() => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  }, [session]);

  useEffect(() => {
    localStorage.setItem(HISTORY_KEY, JSON.stringify(history.slice(0, 100)));
  }, [history]);

  useEffect(() => {
    localStorage.setItem(DRAFTS_KEY, JSON.stringify(drafts));
  }, [drafts]);

  const value = useMemo(
    () => ({
      session,
      setSession,
      result,
      setResult,
      history,
      setHistory,
      drafts,
      setDrafts
    }),
    [session, result, history, drafts]
  );

  return <AppStateContext.Provider value={value}>{children}</AppStateContext.Provider>;
}

export function useAppState() {
  const ctx = useContext(AppStateContext);
  if (!ctx) throw new Error("useAppState must be used within AppStateProvider");
  return ctx;
}


