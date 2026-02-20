import { useMemo, useState } from "react";
import { ExecuteFn, ServiceKey } from "../types";
import { SectionCard } from "./SectionCard";

interface Props {
  execute: ExecuteFn;
}

const SERVICES: { key: ServiceKey; label: string }[] = [
  { key: "gateway", label: "gateway" },
  { key: "order", label: "order-service" },
  { key: "inventory", label: "inventory-service" },
  { key: "query", label: "query-service" },
  { key: "worker", label: "ticket-worker" },
  { key: "nginx", label: "gateway-nginx" }
];

export function HealthSection({ execute }: Props) {
  const [statusMap, setStatusMap] = useState<Record<ServiceKey, string>>({
    gateway: "unknown",
    order: "unknown",
    inventory: "unknown",
    query: "unknown",
    worker: "unknown",
    nginx: "unknown"
  });

  const overall = useMemo(() => {
    const values = Object.values(statusMap);
    if (values.every((v) => v === "ready")) return "ALL_READY";
    if (values.some((v) => v === "down")) return "HAS_DOWN";
    return "PARTIAL";
  }, [statusMap]);

  return (
    <SectionCard title="健康检查" desc="逐服务检查 /healthz，支持一键全检。">
      <div className="toolbar">
        <span className={`badge ${overall.toLowerCase()}`}>总体状态: {overall}</span>
        <button
          onClick={async () => {
            for (const svc of SERVICES) {
              try {
                await execute({
                  label: `${svc.label} /healthz`,
                  service: svc.key,
                  method: "GET",
                  path: "/healthz"
                });
                setStatusMap((m) => ({ ...m, [svc.key]: "ready" }));
              } catch {
                setStatusMap((m) => ({ ...m, [svc.key]: "down" }));
              }
            }
          }}
        >
          一键全检
        </button>
      </div>
      <div className="grid health-grid">
        {SERVICES.map((svc) => (
          <button
            key={svc.key}
            onClick={async () => {
              try {
                await execute({
                  label: `${svc.label} /healthz`,
                  service: svc.key,
                  method: "GET",
                  path: "/healthz"
                });
                setStatusMap((m) => ({ ...m, [svc.key]: "ready" }));
              } catch {
                setStatusMap((m) => ({ ...m, [svc.key]: "down" }));
              }
            }}
          >
            {svc.label}
            <span className={`dot ${statusMap[svc.key]}`} />
          </button>
        ))}
      </div>
    </SectionCard>
  );
}


