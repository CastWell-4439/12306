#!/usr/bin/env sh
set -eu

docker exec -i ticketing-mysql mysql -uroot -proot ticketing < migrations/0001_init.sql
docker exec -i ticketing-mysql mysql -uroot -proot ticketing < migrations/0002_query_readmodel.sql
docker exec -i ticketing-mysql mysql -uroot -proot ticketing < migrations/0003_ticket_outbox.sql

echo "migrations applied"
