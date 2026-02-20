#!/usr/bin/env python3
import argparse
import json
import time
import urllib.request


def call(url: str, payload: dict) -> tuple[int, str]:
    body = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=body, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=3) as resp:
        return resp.status, resp.read().decode("utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="Lightweight load smoke for inventory try-hold")
    parser.add_argument("--url", default="http://127.0.0.1:8082/inventory/try-hold")
    parser.add_argument("--count", type=int, default=100)
    args = parser.parse_args()

    ok = 0
    start = time.time()
    for i in range(args.count):
        payload = {
            "partition_key": "G123|2026-02-11|2nd",
            "hold_id": f"smoke-{i}",
            "qty": 1,
            "capacity": 500,
        }
        try:
            status, _ = call(args.url, payload)
            if 200 <= status < 300:
                ok += 1
        except Exception:
            pass
    took = time.time() - start
    qps = args.count / took if took > 0 else 0
    print(json.dumps({"total": args.count, "ok": ok, "seconds": took, "qps": qps}, ensure_ascii=False))


if __name__ == "__main__":
    main()
