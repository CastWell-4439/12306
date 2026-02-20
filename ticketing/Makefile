.PHONY: up down ps logs init restart clean rebuild health e2e failure-drill

SERVICES ?= gateway order-service inventory-service query-service ticket-worker seat-allocator gateway-nginx

up:
	docker compose up -d --build

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f $(SERVICES)

init:
	docker compose logs migrate topics-init seed

restart:
	docker compose restart $(SERVICES)

clean:
	docker compose down -v --remove-orphans

rebuild:
	docker compose build --no-cache

health:
	@echo "gateway        -> http://127.0.0.1:8080/healthz"
	@echo "order-service  -> http://127.0.0.1:8081/healthz"
	@echo "inventory      -> http://127.0.0.1:8082/healthz"
	@echo "query          -> http://127.0.0.1:8083/healthz"
	@echo "ticket-worker  -> http://127.0.0.1:8084/healthz"
	@echo "gateway-nginx  -> http://127.0.0.1:8088/healthz"

e2e:
	python tools/e2e_order_inventory.py

failure-drill:
	python tools/failure_drill_ticket_outbox.py


