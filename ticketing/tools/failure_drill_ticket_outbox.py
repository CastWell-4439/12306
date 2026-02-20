#!/usr/bin/env python3
import argparse
import json
import subprocess
import time
from datetime import datetime, timezone


def run(cmd: list[str]) -> str:
    out = subprocess.check_output(cmd, stderr=subprocess.STDOUT)
    return out.decode("utf-8", errors="replace").strip()


def mysql_scalar(mysql_container: str, sql: str) -> str:
    return run(
        [
            "docker",
            "exec",
            "-i",
            mysql_container,
            "mysql",
            "-N",
            "-uroot",
            "-proot",
            "ticketing",
            "-e",
            sql,
        ]
    )


def main() -> None:
    parser = argparse.ArgumentParser(description="Failure drill: ticket outbox backlog and recovery when Kafka is down")
    parser.add_argument("--kafka-container", default="ticketing-kafka")
    parser.add_argument("--mysql-container", default="ticketing-mysql")
    parser.add_argument("--timeout-seconds", type=int, default=45)
    args = parser.parse_args()

    ts = int(time.time() * 1000)
    event_id = f"drill-evt-{ts}"
    order_id = f"drill-order-{ts}"
    occurred_at = datetime.now(timezone.utc).isoformat()

    insert_sql = (
        "INSERT INTO ticket_outbox(event_id, aggregate_id, event_type, payload, status, retry_count, next_retry_at) "
        f"VALUES('{event_id}', '{order_id}', 'TicketIssued', "
        f"JSON_OBJECT('event_id','{event_id}','aggregate_id','{order_id}','event_type','TicketIssued',"
        f"'occurred_at','{occurred_at}','payload',JSON_OBJECT('order_id','{order_id}','seat_no','DRILL-1')), "
        "'PENDING', 0, CURRENT_TIMESTAMP);"
    )
    mysql_scalar(args.mysql_container, insert_sql)

    run(["docker", "stop", args.kafka_container])
    time.sleep(3)
    status_down = mysql_scalar(
        args.mysql_container,
        f"SELECT status FROM ticket_outbox WHERE event_id='{event_id}' LIMIT 1;",
    )

    run(["docker", "start", args.kafka_container])
    deadline = time.time() + args.timeout_seconds
    final_status = status_down
    while time.time() < deadline:
        final_status = mysql_scalar(
            args.mysql_container,
            f"SELECT status FROM ticket_outbox WHERE event_id='{event_id}' LIMIT 1;",
        )
        if final_status == "PUBLISHED":
            break
        time.sleep(2)

    result = {
        "event_id": event_id,
        "status_when_kafka_down": status_down,
        "status_after_recovery": final_status,
        "ok": final_status == "PUBLISHED",
    }
    print(json.dumps(result, ensure_ascii=False))
    if not result["ok"]:
        raise SystemExit(1)


if __name__ == "__main__":
    main()


