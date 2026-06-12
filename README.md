# 分布式配置中心 (Distributed Config Center)

> 轻量级、全栈的分布式微服务配置管理中心。
> 支持配置草稿编辑、版本发布、差异对比和历史审计。

---

## 本地环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | ≥ 1.22 (推荐 1.22.5) | 后端 + 示例微服务 |
| Node.js | ≥ 18 | 前端开发与构建 |
| npm | ≥ 9 | 随 Node.js 安装 |
| Bash | ≥ 4.0 | start.sh 和 smoke_test.sh |

> **注意**: Go 需在 `PATH` 中。`start.sh` 会自动探测 `/root/go/bin`、`/usr/local/go/bin` 等常见路径。

---

## 项目结构

```
fse/
├── backend/               # Go + Gin 配置中心后端 (:8080)
│   ├── main.go            # 入口: 种子数据、CORS、路由注册
│   ├── handler/           # 8 个 REST API 处理器
│   └── store/             # 内存存储层 (sync.RWMutex 线程安全)
├── frontend/              # Vue 3 + Element Plus 管理界面 (:5173)
│   └── src/
│       ├── api/           # Axios API 客户端
│       ├── components/    # Vue 组件 (Sidebar/TopBar/ConfigTable/...)
│       └── utils/         # 工具函数 (configDiff)
├── demo-service/          # 示例微服务 — 从配置中心拉取配置 (:3000)
├── tests/
│   └── smoke_test.sh      # 端到端冒烟测试 (17 个检查点)
├── start.sh               # 一键启动脚本
├── ARCHITECTURE.md        # 架构设计文档
├── BUGLOG.md              # Bug 追踪记录
└── LINTLOG.md             # 代码质量记录
```

---

## 一键启动

```bash
bash start.sh
```

`Ctrl+C` 一键停止所有进程。

---

## 分步启动

### 1. 启动后端（配置中心）

```bash
cd backend
go run .
# 监听 :8080
# 启动时自动初始化 3 个种子配置组:
#   order-service/dev, order-service/test, user-service/dev
```

### 2. 启动示例微服务

```bash
cd demo-service
go run .
# 监听 :3000 (从配置中心拉取 server.port，默认 3000)
# 启动时调用 GET /api/configs/order-service/dev/published
# 后端未就绪则重试 5 次后降级启动
```

### 3. 启动前端

```bash
cd frontend
npm install        # 首次运行
npm run dev
# 监听 :5173，Vite 自动代理 /api → localhost:8080
```

---

## 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端管理界面 | http://localhost:5173 | Vue 3 SPA，Linear 风格工作台 |
| 后端 REST API | http://localhost:8080/api/configs | Gin REST API |
| 示例微服务 | http://localhost:3000 | 展示配置拉取效果 |

---

## API 接口

基础地址: `http://localhost:8080`

### 1. 获取所有配置组

```bash
curl http://localhost:8080/api/configs
```

### 2. 获取单个配置组详情（含变更日志）

```bash
curl http://localhost:8080/api/configs/order-service/dev
```

### 3. 获取线上配置（供微服务使用）

```bash
curl http://localhost:8080/api/configs/order-service/dev/published
```

### 4. 新增 / 修改配置项

```bash
curl -X PUT http://localhost:8080/api/configs/order-service/dev/keys/db.url \
  -H "Content-Type: application/json" \
  -d '{"value":"localhost:3306"}'
```

### 5. 删除配置项

```bash
curl -X DELETE http://localhost:8080/api/configs/order-service/dev/keys/db.url
```

### 6. 发布配置

```bash
curl -X POST http://localhost:8080/api/configs/order-service/dev/publish
# DraftData → PublishedData，版本号 +1
```

### 7. 推送（预留）

```bash
curl -X POST http://localhost:8080/api/configs/order-service/dev/push
```

### 8. 监听（预留 SSE）

```bash
curl http://localhost:8080/api/configs/watch
curl -H "Accept: text/event-stream" http://localhost:8080/api/configs/watch
```

### 接口一览

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | `GET` | `/api/configs` | 列出所有配置组 |
| 2 | `GET` | `/api/configs/:service/:env` | 获取单个配置组 + 变更日志 |
| 3 | `GET` | `/api/configs/:service/:env/published` | 获取线上配置（供微服务拉取） |
| 4 | `PUT` | `/api/configs/:service/:env/keys/:key` | 新增或修改配置项 → DraftData |
| 5 | `DELETE` | `/api/configs/:service/:env/keys/:key` | 删除配置项 → 从 DraftData 移除 |
| 6 | `POST` | `/api/configs/:service/:env/publish` | 发布配置 → DraftData 复制到 PublishedData |
| 7 | `POST` | `/api/configs/:service/:env/push` | 推送（预留） |
| 8 | `GET` | `/api/configs/watch` | SSE 监听（预留） |

---

## 配置项 Key 命名约束

Key 出现在 REST URL 路径中，有以下限制：

| 类别 | 字符 | 说明 |
|------|------|------|
| ✅ 推荐 | 字母、数字、`.`、`-`、`_` | 如 `db.url`、`server.port`、`log.level` |
| ⚠️ 可用但不推荐 | 中文、`@`、空格 | 前端 `encodeURIComponent` 处理后可正常使用，但 curl / 调试不便 |
| ❌ 禁止 | `/` | 破坏 Gin 路由 `:service/:env/keys/:key` |
| ❌ 避免 | `?` `#` `&` `%` | URL 保留字符，可能导致路由解析错误 |

**最佳实践**: 使用点分命名法，如 `db.url`、`redis.host`、`log.level`。

---

## 测试

### 后端测试

```bash
cd backend
go test ./... -v           # 运行所有测试 (17 handler + 8 store)
go test ./... -count=1     # 禁用缓存强制重跑
```

### 前端构建

```bash
cd frontend
npm run build
# 产出在 dist/ 目录
```

### 端到端冒烟测试

```bash
bash tests/smoke_test.sh
```

覆盖 6 个阶段 (17 个检查点):
1. 编译 backend + demo-service
2. 启动后端
3. API 基础验证（4 项）
4. 编辑 → 发布流程（5 项）
5. Demo Service 拉取配置（2 项）
6. 结果汇总

---

## 架构概览

```
Frontend (Vue :5173)  ← 5s 轮询 →  Backend (Go+Gin :8080)
                                           │
                                           ├── In-Memory ConfigStore
                                           │   ├── PublishedData (线上)
                                           │   ├── DraftData (草稿)
                                           │   └── ChangeLog (最多 10 条)
                                           │
Demo Service (Go+Gin :3000)  ← 启动时拉取 /published
```

- **草稿/发布分离**: 编辑写入 DraftData，点击 Publish 后复制到 PublishedData
- **版本管理**: 每次发布版本号 +1
- **变更审计**: 每次操作记录到 ChangeLog（每组最多 10 条）
- **线程安全**: `sync.RWMutex` 保护读写

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22.5 + Gin 1.9.1 |
| 前端 | Vue 3 (Composition API) + Element Plus 2.14 + Axios + Vite 8 |
| 示例服务 | Go 1.22.5 + Gin 1.9.1 |
| 测试 | Go `testing` + `httptest` + Bash |
