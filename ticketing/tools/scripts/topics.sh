#!/usr/bin/env sh
set -eu

docker exec -i ticketing-kafka kafka-topics --bootstrap-server 127.0.0.1:9092 --create --if-not-exists --topic order.events --partitions 3 --replication-factor 1
docker exec -i ticketing-kafka kafka-topics --bootstrap-server 127.0.0.1:9092 --create --if-not-exists --topic inventory.events --partitions 3 --replication-factor 1
docker exec -i ticketing-kafka kafka-topics --bootstrap-server 127.0.0.1:9092 --create --if-not-exists --topic ticket.events --partitions 3 --replication-factor 1

echo "topics ensured"
