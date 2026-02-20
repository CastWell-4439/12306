#!/usr/bin/env sh
set -eu

python tools/seed_stations.py
if docker ps --format '{{.Names}}' | grep -q '^ticketing-mysql$'; then
  docker exec -i ticketing-mysql mysql -uroot -proot ticketing < tools/data/stations_seed.sql
fi

echo "seed completed"
