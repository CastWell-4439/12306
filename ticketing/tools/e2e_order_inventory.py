#!/usr/bin/env python3
import argparse
import hashlib
import hmac
import json
import time
import urllib.parse
import urllib.request
from typing import Any, Tuple


def sign(secret: str, order_id: str, provider_txn_id: str, status: str) -> str:
    payload = f"{order_id}|{provider_txn_id}|{status.upper()}".encode("utf-8")
    return hmac.new(secret.encode("utf-8"), payload, hashlib.sha256).hexdigest()


def post_json(url: str, payload: dict[str, Any]) -> Tuple[int, dict[str, Any]]:
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=body, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(req, timeout=5) as resp:
        raw = resp.read().decode("utf-8")
        return resp.status, json.loads(raw) if raw else {}


def get_json(url: str) -> Tuple[int, dict[str, Any]]:
    req = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(req, timeout=5) as resp:
        raw = resp.read().decode("utf-8")
        return resp.status, json.loads(raw) if raw else {}


def main() -> None:
    parser = argparse.ArgumentParser(description="Minimal E2E: order -> reserve(try-hold) -> pay(confirm-hold) -> ticketed")
    parser.add_argument("--order-url", default="http://127.0.0.1:8081")
    parser.add_argument("--query-url", default="http://127.0.0.1:8083")
    parser.add_argument("--partition-key", default="G123|2026-02-11|2nd")
    parser.add_argument("--qty", type=int, default=1)
    parser.add_argument("--capacity", type=int, default=500)
    parser.add_argument("--sign-key", default="dev-payment-sign-key")
    parser.add_argument("--wait-seconds", type=int, default=30)
    args = parser.parse_args()

    idempotency_key = f"e2e-{int(time.time() * 1000)}"
    provider_txn_id = f"txn-{int(time.time() * 1000)}"

    _, created = post_json(
        f"{args.order_url}/orders",
        {"idempotency_key": idempotency_key, "amount_cents": 18800},
    )
    order_id = created.get("OrderID")
    if not order_id:
        raise RuntimeError(f"create order failed: {created}")

    post_json(
        f"{args.order_url}/orders/reserve",
        {
            "order_id": order_id,
            "partition_key": args.partition_key,
            "hold_id": order_id,
            "qty": args.qty,
            "capacity": args.capacity,
        },
    )

    signature = sign(args.sign_key, order_id, provider_txn_id, "SUCCESS")
    post_json(
        f"{args.order_url}/payments/callback",
        {
            "order_id": order_id,
            "provider_txn_id": provider_txn_id,
            "status": "SUCCESS",
            "partition_key": args.partition_key,
            "hold_id": order_id,
            "signature": signature,
        },
    )

    deadline = time.time() + args.wait_seconds
    query_view = {}
    while time.time() < deadline:
        try:
            _, query_view = get_json(
                f"{args.query_url}/query/orders?order_id={urllib.parse.quote(order_id)}"
            )
            if query_view.get("status") in {"PAID", "TICKETED"}:
                break
        except Exception:
            pass
        time.sleep(1)

    _, order_final = get_json(f"{args.order_url}/orders/get?order_id={urllib.parse.quote(order_id)}")
    result = {
        "order_id": order_id,
        "order_status": order_final.get("Status"),
        "query_status": query_view.get("status"),
        "query_seat_no": query_view.get("seat_no"),
        "ok": order_final.get("Status") in {"PAID", "TICKETED"},
    }
    print(json.dumps(result, ensure_ascii=False))
    if not result["ok"]:
        raise SystemExit(1)


if __name__ == "__main__":
    main()



