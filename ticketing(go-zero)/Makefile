.PHONY: up down ps logs init restart clean rebuild health build test gen

SERVICES ?= gateway-api order-api inventory-api query-api ticket-worker

# ======================== Docker 部署 ========================

up:
	docker compose up -d --build

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f $(SERVICES)

init:
	docker compose logs migrate topics-init

restart:
	docker compose restart $(SERVICES)

clean:
	docker compose down -v --remove-orphans

rebuild:
	docker compose build --no-cache

# ======================== 本地开发 ========================

build:
	go build ./...

test:
	go test ./...

gen:
	goctl api go -api spec/api/gateway.api   -dir apps/gateway-api   --style goZero
	goctl api go -api spec/api/order.api     -dir apps/order-api     --style goZero
	goctl api go -api spec/api/inventory.api -dir apps/inventory-api --style goZero
	goctl api go -api spec/api/query.api     -dir apps/query-api     --style goZero

health:
	@echo "gateway-api       -> http://127.0.0.1:8080/healthz"
	@echo "order-api         -> http://127.0.0.1:8081/healthz"
	@echo "inventory-api     -> http://127.0.0.1:8082/healthz"
	@echo "query-api         -> http://127.0.0.1:8083/healthz"
	@echo "ticket-worker     -> http://127.0.0.1:8084/healthz"


