

## 项目入口

- 原版工程（Gin）：`ticketing/`
- go-zero 版工程：`ticketing(go-zero)/`
- 前端控制台：`frontend/`
- 执行规范：`CURSOR_PROMPT_CLEAN_ARCH_v3.md`
- API 文档：`ticketing/docs/openapi/*.yaml`

## Stage 进度

- [x] Stage 1 - Infrastructure Core
- [x] Stage 2 - Inventory Service Core
- [x] Stage 2.5 - Inventory Recoverability (WAL + Snapshot)
- [x] Stage 3 - Order Service
- [x] Stage 4 - Ticket Worker
- [x] Stage 5 - C++ Seat Allocator
- [x] Stage 6 - Query Service
- [x] Stage 7 - Python + Shell Tooling
- [x] Stage 8 - go-zero 框架迁移

## go-zero 迁移说明


| 维度 | 说明 |
|------|------|
| 目录 | `ticketing(go-zero)/` — 与原项目并行，不删除旧代码 |
| 骨架生成 | `goctl api go` 从 `spec/api/*.api` 生成 handler/logic/types |
| 核心逻辑 | `pkg/core/` — 从原 `internal/` 迁出，业务代码零改动 |
| 基础设施 | `pkg/infra/` — kafka/mysql/redis/仓储层 |
| HTTP 路径 | 与原项目 100% 一致，前端无感 |
| 内置能力 | Timeout / MaxConns / Prometheus / Telemetry 通过 yaml 一行启用 |
| 日志 | Logic 层内嵌 `logx.Logger`，自动携带 traceId |
| 测试 | 4 个核心测试套件全部通过（partition / order / query / ticket） |

### go-zero 服务清单

| 服务 | 入口 | 端口 | .api 定义 |
|------|------|------|-----------|
| gateway-api | `apps/gateway-api/gateway.go` | 8080 | `spec/api/gateway.api` |
| order-api | `apps/order-api/order.go` | 8081 | `spec/api/order.api` |
| inventory-api | `apps/inventory-api/inventory.go` | 8082 | `spec/api/inventory.api` |
| query-api | `apps/query-api/query.go` | 8083 | `spec/api/query.api` |
| ticket-worker | `apps/ticket-worker/worker.go` | 8084 | — (后台 Worker) |

### go-zero 版快速验证

```bash
cd "ticketing(go-zero)"
go build ./...
go test ./...
```

## Swagger / OpenAPI

- `ticketing/docs/openapi/order-service.openapi.yaml`
- `ticketing/docs/openapi/inventory-service.openapi.yaml`
- `ticketing/docs/openapi/query-service.openapi.yaml`
- `ticketing/docs/openapi/ticket-worker.openapi.yaml`
- `ticketing/docs/openapi/gateway.openapi.yaml`
- 说明：`ticketing/docs/openapi/README.md`

## 部署（原版 Gin）

### 前置条件

- 已安装 Docker Engine（建议 24+）
- 已安装 Docker Compose 插件（`docker compose` 可用）
- 服务器能访问 Docker Hub（首次会拉镜像）

### 1) 启动

```bash
cd ticketing
docker compose up -d --build
```

这条命令会自动完成：

- 基础依赖启动（MySQL/Redis/Kafka）
- 初始化任务执行（`migrate` / `topics-init` / `seed`）
- 业务服务启动（gateway/order/inventory/query/ticket-worker）
- 监控组件启动（Prometheus/Grafana）
- 前端控制台启动（frontend，容器化部署）

### 2) 检查容器状态

```bash
cd ticketing
docker compose ps
```

### 3) 查看初始化任务日志

```bash
cd ticketing
docker compose logs migrate topics-init seed
```

### 4) （可选）在"只有 Docker"的机器上做健康检查

如果服务器没有 `curl`，可以用临时 curl 容器验证：

```bash
docker run --rm --network host curlimages/curl:8.9.1 http://127.0.0.1:8080/healthz
docker run --rm --network host curlimages/curl:8.9.1 http://127.0.0.1:8081/healthz
docker run --rm --network host curlimages/curl:8.9.1 http://127.0.0.1:8082/healthz
docker run --rm --network host curlimages/curl:8.9.1 http://127.0.0.1:8083/healthz
docker run --rm --network host curlimages/curl:8.9.1 http://127.0.0.1:8084/healthz
```

> 若你的环境不支持 `--network host`，可改用宿主机自带工具或从外部机器访问开放端口。

### 5) 停止与清理

```bash
cd ticketing
docker compose down
```

如需连数据卷一起清理：

```bash
cd ticketing
docker compose down -v --remove-orphans
```

## 快速启动（通用）

```bash
cd ticketing
docker compose up -d --build
```

说明：`migrate/topics-init/seed` 会作为初始化任务自动执行。

## 统一命令入口（可选）

- `ticketing/Makefile`：适合 Linux/macOS（需要 `make`）
- `ticketing/justfile`：适合 Windows/Linux（需要 `just`）

示例：

```bash
cd ticketing
make up
make ps
make logs
```

```bash
cd ticketing
just up
just ps
just logs
```

## 关键验证命令

```bash
cd ticketing
go test ./...
```

```bash
cd ticketing
python tools/e2e_order_inventory.py
```

```bash
cd ticketing
python tools/failure_drill_ticket_outbox.py
```

说明：

- `e2e_order_inventory.py`：验证下单 -> 预留占座 -> 支付回调 -> 订单终态闭环。
- `failure_drill_ticket_outbox.py`：验证 Kafka 故障下 `ticket_outbox` 堆积与恢复补发。

## 常用访问地址

- Gateway health: `http://127.0.0.1:8080/healthz`
- Order service: `http://127.0.0.1:8081`
- Inventory service: `http://127.0.0.1:8082`
- Query service: `http://127.0.0.1:8083`
- Ticket worker health: `http://127.0.0.1:8084/healthz`
- OpenResty gateway: `http://127.0.0.1:8088/healthz`
- Frontend: `http://127.0.0.1:5173`
- Prometheus: `http://127.0.0.1:9090`
- Grafana: `http://127.0.0.1:3000`（默认 `admin/admin`）

## 监控（Prometheus + Grafana）

- 每个 Go 服务已暴露 `GET /metrics`：
  - `gateway:8080/metrics`
  - `order-service:8081/metrics`
  - `inventory-service:8082/metrics`
  - `query-service:8083/metrics`
  - `ticket-worker:8084/metrics`
- 内置 Prometheus 抓取配置：`ticketing/deployments/observability/prometheus.yml`
- 指标包含：
  - `http_requests_total`
  - `http_request_duration_seconds`
  - `http_in_flight_requests`
- go-zero 版额外暴露独立 Prometheus 端口（9080–9084），通过 yaml `Prometheus` 配置启用

## 前端控制台

- 路径：`frontend/`
- 作用：对接当前后端 API（健康检查、订单、库存、查询）
- 服务端部署（仅 Docker）：
  - 随 `docker compose up -d --build` 自动启动
  - 访问 `http://127.0.0.1:5173`
- 本地开发模式（可选）：
  - `cd frontend`
  - `npm install`
  - `npm run dev`

## 排障建议

- 查看所有容器状态：`docker compose ps`
- 查看初始化任务日志：`docker compose logs migrate topics-init seed`
- 查看服务日志：`docker compose logs -f order-service inventory-service query-service ticket-worker`
