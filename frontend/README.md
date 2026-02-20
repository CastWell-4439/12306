# Ticketing Frontend

Frontend console for the current ticketing backend APIs.

## 功能分块

- 健康检查模块
  - 单服务健康检查
  - 一键全检
  - 总体状态徽章（ALL_READY/HAS_DOWN/PARTIAL）
- 订单模块
  - 创建订单
  - 预占订单
  - 支付回调
  - 查询订单
  - 自动提取 `OrderID` 回填到共享字段
- 库存模块
  - TryHold
  - ReleaseHold
  - ConfirmHold
  - 查询可用量
- 查询模块
  - 查询 `query-service` 读模型
  - 支持“使用当前 order_id”快捷填充
- 流程编排模块
  - 一键执行订单主流程（创建 -> 预占 -> 支付 -> 查询）
  - 一键执行库存流程（TryHold -> ConfirmHold -> 查询可用量）
- 结果面板
  - 展示最近一次响应 JSON
  - 展示 HTTP 状态码与耗时
- 请求历史模块
  - 记录调用顺序、状态码、耗时、服务归属
  - 一键清空 + 一键重放

## 产品化能力（非 UI 设计）

- 路由化结构（React Router）
  - `/login`
  - `/app/dashboard`
  - `/app/booking`
  - `/app/inventory`
  - `/app/orders`
- 受保护路由（登录态守卫）
- 全局状态（Context）：
  - session
  - 请求历史
  - 最近结果
  - 业务草稿（localStorage 持久化）
- 统一请求执行管道：
  - 请求耗时统计
  - 统一错误处理
  - 历史记录入库
- 订单轮询能力（Booking 页可开启 3s 轮询）
- 请求重放能力（History 面板）

## Run

```bash
cd frontend
npm install
npm run dev
```

Open `http://127.0.0.1:5173`.

## Notes

- Vite dev server proxies are pre-configured:
  - `/api/order` -> `http://127.0.0.1:8081`
  - `/api/inventory` -> `http://127.0.0.1:8082`
  - `/api/query` -> `http://127.0.0.1:8083`
  - `/api/gateway` -> `http://127.0.0.1:8080`
  - `/api/worker` -> `http://127.0.0.1:8084`
  - `/api/nginx` -> `http://127.0.0.1:8088`
- Start backend stack first.

## 目录结构

- `src/App.tsx`：顶层路由
- `src/state/AppState.tsx`：全局状态与本地持久化
- `src/hooks/useExecutor.ts`：统一执行器+重放
- `src/pages/`：页面层（login/dashboard/booking/inventory/orders）
- `src/components/`：业务功能模块组件
- `src/api.ts`：HTTP 请求封装
- `src/types.ts`：共享类型定义


