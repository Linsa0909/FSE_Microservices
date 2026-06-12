# 分布式配置中心 (Distributed Config Center)

> 轻量级全栈配置中心 MVP，含 Go 后端、Vue 3 前端、多域仿真微服务平台。
> 支持草稿/发布分离、环境隔离、版本管理、变更审计。
<<<<<<< HEAD
=======

---

## 快速开始

```bash
git clone <repo-url> && cd fse && bash start.sh
```

首次运行会自动安装前端依赖（`npm install`），后续启动秒级完成。`Ctrl+C` 一键停止全部 6 个进程。
>>>>>>> a77b208 (docs: 交付收尾 — 文档终稿 + start.sh 开箱即用 + 同网段访问)

---

## 本地环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | ≥ 1.22 | 后端 + 示例微服务 |
| Node.js | ≥ 18 | 前端开发与构建 |
| npm | ≥ 9 | 随 Node.js 安装 |
| Python 3 | ≥ 3.6 | 冒烟测试脚本使用 `json.tool` |

> Go 需在 `PATH` 中。`start.sh` 自动探测 `/root/go/bin`、`/usr/local/go/bin` 等路径。

---

## 项目结构

```
fse/
├── backend/                     # Go + Gin 配置中心后端 (:8080)
│   ├── main.go                  # 入口: 10 组种子配置 + CORS + 8 路由
│   ├── handler/config.go        # 8 个 REST API handler
│   ├── handler/handler_test.go  # 21 个集成测试
│   ├── store/store.go           # 内存存储 (sync.RWMutex + Draft/Published)
│   └── store/store_test.go      # 12 个单元测试
│
├── frontend/                    # Vue 3 + Element Plus 管理界面 (:5173)
│   └── src/
│       ├── App.vue              # 根组件: Dashboard + 3 Tab + 帮助
│       ├── api/configService.js # Axios API 封装
│       ├── utils/configDiff.js  # deepEqual + hasDraft
│       └── components/          # 9 个 Vue 组件
│           ├── Sidebar.vue      # 服务/环境/状态三阶过滤
│           ├── TopBar.vue       # 搜索框
│           ├── ConfigTable.vue  # 配置组列表
│           ├── ConfigDrawer.vue # Diff 编辑面板
│           ├── ChangeLog.vue    # 变更时间线
│           ├── StatusBadge.vue  # 状态徽章
│           ├── ServiceStatus.vue # 微服务健康探测
│           └── EnvironmentPanel.vue # 环境与推送
│
├── demo-service/                # 多域仿真微服务平台
│   ├── main.go                  # SERVICE_TYPE 分发器 + CORS
│   ├── pkg/configclient/        # 共享 ConfigClient SDK
│   │   ├── client.go            # PullConfig + GetString/Int/Float/Enum
│   │   ├── diagnostics.go       # 通用配置诊断
│   │   └── client_test.go       # 9 个 SDK 测试
│   └── internal/
│       ├── radar/               # 🔴 火控雷达 (:3001)
│       ├── sensor/              # 📷 光电传感器 (:3002)
│       └── navigation/          # 🚢 船舶航海 (:3003)
│
├── tests/
│   └── smoke_test.sh            # 端到端冒烟测试 (22 检查点)
├── start.sh                     # 一键启动 6 进程
├── ARCHITECTURE.md              # 架构设计文档
├── DEMO_SERVICES.md             # 领域服务说明文档
├── AI_PROMPTS.md                # AI 协同研发报告
├── BUGLOG.md                    # Bug 追踪记录
└── LINTLOG.md                   # 代码质量记录
```

---

## 一键启动

```bash
bash start.sh
```

启动 6 个进程，`Ctrl+C` 一键停止。

---

## 分步启动

### 1. 启动后端（配置中心）

```bash
cd backend && go run .
# 监听 :8080
# 启动时初始化 10 个种子配置组:
#   通用: order-service/dev, order-service/test, user-service/dev, user-service/prod
#   领域: radar-service/dev+prod, sensor-service/dev+prod, navigation-service/dev+prod
```

### 2. 启动领域仿真服务（3 个）

```bash
cd demo-service

# 火控雷达 :3001
SERVICE_TYPE=radar CONFIG_SERVICE=radar-service CONFIG_ENV=dev RADAR_PORT=3001 go run .

# 光电传感器 :3002
SERVICE_TYPE=sensor CONFIG_SERVICE=sensor-service CONFIG_ENV=dev SENSOR_PORT=3002 go run .

# 船舶航海 :3003
SERVICE_TYPE=navigation CONFIG_SERVICE=navigation-service CONFIG_ENV=dev NAV_PORT=3003 go run .

# 默认配置回显 :3000 (可选)
SERVICE_TYPE=default CONFIG_SERVICE=order-service CONFIG_ENV=dev PORT=3000 go run .
```

### 3. 启动前端

```bash
cd frontend
npm install   # 首次运行
npm run dev   # 监听 :5173，Vite 代理 /api→:8080
```

---

## 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 前端管理界面 | http://localhost:5173 | Vue 3 SPA，Linear 风格工作台 |
| 后端 REST API | http://localhost:8080/api/configs | 8 个 REST 端点 |
| 📦 默认示例服务 | http://localhost:3000 | 配置回显 |
| 🔴 火控雷达 | http://localhost:3001/radar/status | 目标模拟 + 状态机 |
| 📷 光电传感器 | http://localhost:3002/sensor/status | 帧生成 + 检测 |
| 🚢 船舶航海 | http://localhost:3003/nav/status | GPS 导航 + 航点 |

---

## 示例微服务验证地址说明

> **含义**：启动每个 demo-service 后，通过以下 URL 验证它**确实从配置中心拉取了配置并正常运行**。

### 健康检查（所有服务通用）

| 服务 | 验证地址 | 预期响应 |
|------|----------|----------|
| 默认 | http://localhost:3000/health | `{"status":"ok","config_from_center":true}` |
| 雷达 | http://localhost:3001/health | `{"status":"ok","config_from_center":true}` |
| 传感器 | http://localhost:3002/health | `{"status":"ok","config_from_center":true}` |
| 航海 | http://localhost:3003/health | `{"status":"ok","config_from_center":true}` |

> `config_from_center=true` 证明服务成功从配置中心拉取了 PublishedData。
> `config_from_center=false` 表示降级启动（配置中心不可达，使用默认配置）。

### 领域功能验证

```bash
# 雷达：查看目标列表和扫描状态
curl http://localhost:3001/radar/status
# → active_targets: 32, mode: "search", config.scan_rate_hz: 60

# 传感器：获取最新一帧检测结果
curl http://localhost:3002/sensor/frame
# → frame_id, detections[{type, confidence, bbox}]

# 航海：查看当前 GPS 位置和航线
curl http://localhost:3003/nav/status
# → current_position, heading_deg, eta, waypoint_progress
```

### 配置驱动验证

1. 在前端管理界面修改 `radar.mode` 为 `engage`
2. 发布配置
3. 雷达服务**无需重启**，调用 `POST /radar/mode` 动态切换，或重启后读取新配置
4. 验证：`curl http://localhost:3001/radar/status` → mode 变为 `engage`

---

## API 接口

基础地址: `http://localhost:8080`

| # | Method | Path | 说明 |
|---|--------|------|------|
| 1 | `GET` | `/api/configs` | 列出所有配置组 |
| 2 | `GET` | `/api/configs/:service/:env` | 获取单个配置组 + 变更日志 |
| 3 | `GET` | `/api/configs/:service/:env/published` | 获取线上配置（供微服务拉取） |
| 4 | `PUT` | `/api/configs/:service/:env/keys/:key` | 新增或修改配置项 → DraftData |
| 5 | `DELETE` | `/api/configs/:service/:env/keys/:key` | 删除配置项 |
| 6 | `POST` | `/api/configs/:service/:env/publish` | 发布配置 → PublishedData + 版本号 +1 |
| 7 | `POST` | `/api/configs/:service/:env/push` | 推送（预留） |
| 8 | `GET` | `/api/configs/watch` | SSE 监听（预留） |

### curl 示例

```bash
# 查看所有配置组
curl http://localhost:8080/api/configs

# 查看雷达服务配置详情（含变更日志）
curl http://localhost:8080/api/configs/radar-service/dev

# 微服务拉取线上配置
curl http://localhost:8080/api/configs/radar-service/dev/published

# 修改配置项
curl -X PUT http://localhost:8080/api/configs/radar-service/dev/keys/radar.mode \
  -H "Content-Type: application/json" -d '{"value":"track"}'

# 发布
curl -X POST http://localhost:8080/api/configs/radar-service/dev/publish
```

---

## 配置项 Key 命名约束

| 类别 | 字符 | 示例 |
|------|------|------|
| ✅ 推荐 | 字母、数字、`.`、`-`、`_` | `db.url`、`radar.scan_rate_hz` |
| ⚠️ 可用 | 中文、`@`、空格 | `配置项.名称`（需 encodeURIComponent） |
| ❌ 禁止 | `/` | 破坏 Gin `:service/:env/keys/:key` 路由 |

---

## 测试

### 后端

```bash
cd backend
go test ./... -v -count=1
# store: 12 tests, handler: 21 tests
```

### 前端

```bash
cd frontend && npm run build
```

### SDK

```bash
cd demo-service
go test ./pkg/configclient/ -v -count=1
# 9 tests: GetString/Int/Float/Enum + PullConfig 重试/失败 + Diagnostics
```

### 冒烟测试

```bash
bash tests/smoke_test.sh
# 22 检查点: 编译 → 10配置组验证 → 编辑发布 → 3领域服务启动+health+API → 模式切换
```

---

## 核心概念

| 概念 | 定义 | 示例 |
|------|------|------|
| **配置组** | `service:env` 组合，配置隔离的最小单元 | `radar-service:dev`（共 10 组） |
| **配置项** | 配置组内的单对 key-value | `radar.scan_rate_hz=60`（radar 组有 5 项） |
| **草稿** | DraftData，编辑但未发布 | 修改后立即生效于前端，微服务不可见 |
| **发布** | DraftData → PublishedData，版本号 +1 | 微服务下次拉取 /published 获得新值 |

---

## 架构概览

```
Frontend (Vue :5173)  ← 5s 轮询 →  Backend (Go :8080)
                                          │
                                          ├── ConfigStore (内存)
                                          │   ├── PublishedData (线上)
                                          │   ├── DraftData (草稿)
                                          │   └── ChangeLog (≤10条)
                                          │
Radar    :3001 ─┐                         │
Sensor   :3002 ─┼─ 启动拉取 /published ───┘
Nav      :3003 ─┘   降级启动 (5次重试)
Default  :3000 ─┘
```

- **草稿/发布分离**：编辑 ≠ 生效，Publish 后微服务方可读取
- **版本管理**：semver 补丁号递增
- **变更审计**：自动记录操作类型+时间+版本号
- **多域仿真**：火控雷达/光电传感器/船舶航海，各有独立配置契约和业务逻辑
