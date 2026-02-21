# Ticketing Architecture Overview

## Runtime Topology

```mermaid
flowchart LR
  FE[Frontend] --> GW[gateway-nginx]
  GW --> ORD[order-service]
  GW --> INV[inventory-service]
  GW --> QRY[query-service]

  ORD --> DB[(MySQL)]
  ORD --> K[(Kafka order.events)]

  INV --> R[(Redis hold TTL)]
  INV --> DB
  INV --> K2[(Kafka inventory.events)]

  TKW[ticket-worker] --> DB
  TKW --> TO[(ticket_outbox)]
  TO -->|publisher loop| K3[(Kafka ticket.events)]

  QRY --> DB
  QRY --> R
  QRY -->|consume| K
  QRY -->|consume| K3
```

## Consistency Rules

- Inventory mutations are serialized by partitioned actor model.
- Memory state must not advance without durable WAL/outbox acceptance.
- Query consumer commits offset only after DB transaction succeeds.
- Ticket events are produced from `ticket_outbox` to guarantee eventual delivery.



