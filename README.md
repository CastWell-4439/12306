# 12306 售票系统

基于微服务架构的高并发火车票售票系统，涵盖下单、库存锁定、支付、出票全流程。

## 项目结构

```
12306/
├── ticketing/              # 后端微服务（Go，Gin 版）
├── ticketing(go-zero)/     # 后端微服务（Go，go-zero 版）
├── frontend/               # 前端控制台（Vue 3 + TypeScript）
└── China-rail-way-stations-data-main/  # 车站基础数据
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Gin（原版）/ go-zero（新版，goctl 生成骨架） |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | Apache Kafka (KRaft) |
| 座位分配 | C++ gRPC 服务（seat-allocator） |
| 前端 | Vue 3 + Vite + TypeScript |
| 监控 | Prometheus + Grafana |
| 容器化 | Docker Compose 一键部署 |

## 服务清单

| 服务 | 端口 | 职责 |
|------|------|------|
| gateway | 8080 | API 网关、健康检查 |
| order-service | 8081 | 订单创建、预留、支付回调、取消 |
| inventory-service | 8082 | 库存锁定（WAL + Snapshot 恢复）、TTL 自动释放 |
| query-service | 8083 | 订单查询读模型（CQRS） |
| ticket-worker | 8084 | 消费 OrderPaid 事件 → 分配座位 → 出票 |
| seat-allocator | 50051 | C++ gRPC 座位分配（可选，默认 mock） |
| frontend | 5173 | Web 控制台 |
| prometheus | 9090 | 指标采集 |
| grafana | 3000 | 监控面板（默认 admin/admin） |

## 快速启动

两套后端共享同一数据库和 Kafka，端口完全一致，**任选其一启动**即可。

### Gin 版（原版）

```bash
cd ticketing
docker compose up -d --build     # 一键启动全部
```

或使用 make / just：

```bash
cd ticketing
make up          # 启动
make ps          # 查看状态
make logs        # 查看日志
make health      # 打印健康检查地址
make down        # 停止
make clean       # 停止并清除数据卷
```

```bash
cd ticketing
just up / just ps / just logs / just down / just clean
```

### go-zero 版

```bash
cd "ticketing(go-zero)"
docker compose up -d --build     # 一键启动全部
```

或使用 make / just：

```bash
cd "ticketing(go-zero)"
make up          # 启动
make ps          # 查看状态
make logs        # 查看日志
make health      # 打印健康检查地址
make gen         # goctl 重新生成骨架代码
make down        # 停止
make clean       # 停止并清除数据卷
```

```bash
cd "ticketing(go-zero)"
just up / just ps / just logs / just gen / just down / just clean
```

### 验证服务

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8082/healthz
curl http://127.0.0.1:8083/healthz
curl http://127.0.0.1:8084/healthz
```

打开浏览器访问：

- 前端控制台：http://127.0.0.1:5173
- Grafana 面板：http://127.0.0.1:3000

### 停止

```bash
docker compose down            # 保留数据卷
docker compose down -v         # 连数据卷一起清除
```

## API 文档

OpenAPI 3.0 规范位于 `ticketing/docs/openapi/`：

| 文件 | 对应服务 |
|------|----------|
| `order-service.openapi.yaml` | 订单：创建 / 预留 / 支付 / 取消 / 查询 |
| `inventory-service.openapi.yaml` | 库存：try-hold / release / confirm / availability |
| `query-service.openapi.yaml` | 查询：订单读模型 |
| `gateway.openapi.yaml` | 网关：healthz / readyz |
| `ticket-worker.openapi.yaml` | 出票 Worker：healthz / readyz |

可使用 [Swagger Editor](https://editor.swagger.io) 导入查看。

## 核心业务流程

```
用户下单 → 创建订单(INIT)
       → 预留库存(RESERVED) ── 失败回滚释放
       → 支付回调(PAID) ── 确认库存扣减
       → 出票(TICKETED) ── seat-allocator 分配座位
       → 查询服务异步更新读模型
```

关键设计：

- **Outbox 模式**：订单状态变更通过 outbox 表可靠投递到 Kafka
- **WAL + Snapshot**：库存服务基于 Write-Ahead Log 保证崩溃恢复
- **TTL 自动释放**：预留超时后 Redis delay queue 自动释放库存
- **幂等性**：所有写接口均支持幂等重试

## 两套后端对比

| | Gin 版 (`ticketing/`) | go-zero 版 (`ticketing(go-zero)/`) |
|---|---|---|
| 框架 | Gin + 手写路由 | go-zero + goctl 生成骨架 |
| 内置能力 | 手动实现 | 超时/限流/Prometheus/链路追踪 yaml 配置即启用 |
| 日志 | slog | logx（自动携带 traceId） |
| 一键部署 | `make up` | `make up` |
| 代码生成 | — | `make gen`（goctl 从 .api 重新生成） |
| 本地编译 | `go build ./...` | `make build` |
| 单元测试 | `go test ./...` | `make test` |

## 本地开发

### 后端（需要本地基础设施）

```bash
# Gin 版
cd ticketing
docker compose up -d mysql redis kafka   # 只启动依赖
go run cmd/order-service/main.go         # 启动单个服务
go test ./...

# go-zero 版
cd "ticketing(go-zero)"
docker compose up -d mysql redis kafka   # 只启动依赖
go run apps/order-api/order.go           # 启动单个服务
make test
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

### 端到端验证

```bash
cd ticketing
python tools/e2e_order_inventory.py              # 全流程闭环测试
python tools/failure_drill_ticket_outbox.py       # 故障恢复演练
```

## 监控

每个服务暴露 `GET /metrics`，Prometheus 自动抓取。

内置指标：`http_requests_total` / `http_request_duration_seconds` / `http_in_flight_requests`

配置文件：`ticketing/deployments/observability/prometheus.yml`

## 排障

```bash
docker compose ps                                                        # 容器状态
docker compose logs migrate topics-init                                  # 初始化日志
docker compose logs -f order-service inventory-service query-service     # 服务日志（Gin 版）
docker compose logs -f order-api inventory-api query-api                 # 服务日志（go-zero 版）
```
