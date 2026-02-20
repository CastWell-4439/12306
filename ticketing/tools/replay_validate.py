#!/usr/bin/env python3
import argparse
import json


def unwrap_payload(event: dict) -> dict:
    payload = event.get("payload", {})
    if isinstance(payload, dict) and isinstance(payload.get("payload"), dict):
        return payload["payload"]
    if isinstance(payload, dict):
        return payload
    return {}


def main() -> None:
    parser = argparse.ArgumentParser(description="Replay validation from exported events jsonl")
    parser.add_argument("--input", required=True, help="jsonl file path")
    args = parser.parse_args()

    hold_qty = {}
    capacity = {}
    held_total = {}
    confirmed_total = {}

    paid_orders = set()
    ticketed_orders = set()
    duplicate_event_ids = set()
    seen_event_ids = set()
    invalid_events = []

    with open(args.input, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            ev = json.loads(line)
            event_type = ev.get("event_type", "")
            payload = unwrap_payload(ev)
            event_id = ev.get("event_id", "")
            if event_id:
                if event_id in seen_event_ids:
                    duplicate_event_ids.add(event_id)
                seen_event_ids.add(event_id)

            if event_type == "hold_created":
                key = payload.get("partition_key", "unknown")
                hold_id = payload.get("hold_id", "")
                qty = int(payload.get("qty", 0))
                cap = int(payload.get("capacity", 0))
                if qty <= 0 or not hold_id:
                    invalid_events.append({"type": event_type, "reason": "invalid hold_created payload", "event_id": event_id})
                    continue
                if cap > 0 and key not in capacity:
                    capacity[key] = cap
                hold_qty[hold_id] = qty
                held_total[key] = held_total.get(key, 0) + qty
            elif event_type == "hold_released":
                key = payload.get("partition_key", "unknown")
                hold_id = payload.get("hold_id", "")
                qty = int(payload.get("qty", 0))
                prev = hold_qty.pop(hold_id, qty)
                held_total[key] = held_total.get(key, 0) - prev
            elif event_type == "OrderPaid":
                paid_orders.add(ev.get("aggregate_id", ""))
            elif event_type == "TicketIssued":
                ticketed_orders.add(ev.get("aggregate_id", ""))
            elif event_type == "hold_confirmed":
                key = payload.get("partition_key", "unknown")
                hold_id = payload.get("hold_id", "")
                qty = int(payload.get("qty", 0))
                prev = hold_qty.pop(hold_id, qty)
                held_total[key] = held_total.get(key, 0) - prev
                confirmed_total[key] = confirmed_total.get(key, 0) + prev

    negative_held_keys = sorted([k for k, v in held_total.items() if v < 0])
    over_capacity_keys = sorted(
        [k for k, held in held_total.items() if capacity.get(k, 0) > 0 and held + confirmed_total.get(k, 0) > capacity[k]]
    )
    invalid_ticketed = sorted([oid for oid in ticketed_orders if oid not in paid_orders])

    result = {
        "negative_held_keys": negative_held_keys,
        "over_capacity_keys": over_capacity_keys,
        "ticketed_without_paid": invalid_ticketed,
        "duplicate_event_ids": sorted(list(duplicate_event_ids)),
        "invalid_events": invalid_events,
        "ok": len(negative_held_keys) == 0
        and len(over_capacity_keys) == 0
        and len(invalid_ticketed) == 0
        and len(duplicate_event_ids) == 0
        and len(invalid_events) == 0,
    }
    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
