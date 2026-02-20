#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def main() -> None:
    parser = argparse.ArgumentParser(description="Offline station seed preview")
    parser.add_argument("--input", default="tools/data/stations.json")
    parser.add_argument("--output", default="tools/data/stations_seed.sql")
    args = parser.parse_args()

    stations = json.loads(Path(args.input).read_text(encoding="utf-8-sig"))
    lines = [
        "CREATE TABLE IF NOT EXISTS stations (",
        "  station_code VARCHAR(16) PRIMARY KEY,",
        "  station_name VARCHAR(64) NOT NULL",
        ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;",
        "",
    ]
    for item in stations:
        code = item["station_code"]
        name = item["station_name"].replace("'", "''")
        lines.append(
            f"INSERT INTO stations(station_code, station_name) VALUES('{code}', '{name}') ON DUPLICATE KEY UPDATE station_name=VALUES(station_name);"
        )

    Path(args.output).write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"generated: {args.output}, rows={len(stations)}")


if __name__ == "__main__":
    main()
