#!/usr/bin/env python3
import argparse
import csv
import hashlib
import json
from pathlib import Path
from typing import Iterable


def _escape_sql(value: str) -> str:
    return value.replace("'", "''")


def _deterministic_code(name: str, province: str, city: str) -> str:
    # station_code VARCHAR(16), use a stable hash key for CSV sources without telecode.
    raw = f"{name}|{province}|{city}".encode("utf-8")
    return f"S{hashlib.md5(raw).hexdigest()[:12].upper()}"


def _load_json(path: Path) -> list[dict[str, str]]:
    raw = json.loads(path.read_text(encoding="utf-8-sig"))
    rows: list[dict[str, str]] = []
    for item in raw:
        code = str(item.get("station_code", "")).strip()
        name = str(item.get("station_name", "")).strip()
        if not code or not name:
            continue
        rows.append({"station_code": code[:16], "station_name": name})
    return rows


def _iter_csv_rows(path: Path, encodings: Iterable[str]) -> tuple[list[str], list[list[str]]]:
    last_error: Exception | None = None
    for enc in encodings:
        try:
            with path.open("r", encoding=enc, newline="") as f:
                reader = csv.reader(f)
                header = next(reader, [])
                rows = [row for row in reader if row]
                return header, rows
        except Exception as exc:
            last_error = exc
    if last_error is not None:
        raise last_error
    raise RuntimeError("unable to read CSV")


def _load_csv(path: Path) -> list[dict[str, str]]:
    _, rows = _iter_csv_rows(path, ("utf-8-sig", "utf-8", "gb18030", "gbk"))

    result: list[dict[str, str]] = []
    seen: set[str] = set()
    for row in rows:
        # Dataset layout: index, station_name, ..., province, city, ...
        if len(row) < 2:
            continue
        name = row[1].strip()
        if not name:
            continue
        province = row[6].strip() if len(row) > 6 else ""
        city = row[7].strip() if len(row) > 7 else ""
        code = _deterministic_code(name=name, province=province, city=city)
        if code in seen:
            continue
        seen.add(code)
        result.append({"station_code": code, "station_name": name})
    return result


def _load_stations(path: Path) -> list[dict[str, str]]:
    suffix = path.suffix.lower()
    if suffix == ".json":
        return _load_json(path)
    if suffix == ".csv":
        return _load_csv(path)
    raise ValueError(f"unsupported input format: {path}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Offline station seed preview")
    parser.add_argument("--input", default="tools/data/stations.json")
    parser.add_argument("--output", default="tools/data/stations_seed.sql")
    args = parser.parse_args()

    stations = _load_stations(Path(args.input))
    lines = [
        "CREATE TABLE IF NOT EXISTS stations (",
        "  station_code VARCHAR(16) PRIMARY KEY,",
        "  station_name VARCHAR(64) NOT NULL",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
        "",
    ]
    for item in stations:
        code = _escape_sql(item["station_code"])
        name = _escape_sql(item["station_name"])
        lines.append(
            f"INSERT INTO stations(station_code, station_name) VALUES('{code}', '{name}') ON DUPLICATE KEY UPDATE station_name=VALUES(station_name);"
        )

    Path(args.output).write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"generated: {args.output}, rows={len(stations)}, source={args.input}")


if __name__ == "__main__":
    main()
