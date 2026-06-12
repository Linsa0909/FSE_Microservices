# 架构设计文档 (ARCHITECTURE.md)

## 一、项目概述

分布式配置中心 — 轻量级 MVP，模拟 Nacos/Apollo 极简核心能力。支持配置集中存储、草稿/发布分离、环境隔离、前端可视化管理、示例微服务拉取配置。

---

## 二、整体架构拓扑

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                       │
│  ┌──────────────────┐         5s 轮询        ┌────────────────────┐  │
│  │  前端管理界面      │ ◄─────────────────── │  配置中心后端       │  │
│  │  Vue 3 + Element  │                       │  Go + Gin :8080    │  │
│  │  Plus :5173       │ ────────────────────► │                    │  │
│  │                   │  PUT/POST/DELETE       │  ┌──────────────┐  │  │
│  │  ┌─────────────┐  │  (新增/修改/删除/发布)   │  │ ConfigStore  │  │  │
│  │  │ Dashboard   │  │                       │  │ (sync.RWMutex│  │  │
│  │  │ + 3 Tab页   │  │                       │  │  内存存储)    │  │  │
│  │  ├─配置管理    │  │                       │  │              │  │  │
│  │  ├─微服务状态  │  │                       │  │ PublishedData │  │  │
│  │  └─环境与推送  │  │                       │  │ DraftData     │  │  │
│  └──────────────────┘                       │  │ ChangeLog(10) │  │  │
│                                              │  └──────────────┘  │  │
│                                              │    10 配置组        │  │
│                                              └────────────────────┘  │
│                                                 │                    │
│  ┌──────────────────────────────────────────────┼───────────────┐   │
│  │  demo-service (单 binary, SERVICE_TYPE 切换)  │ GET/published │   │
│  │                                              │               │   │
│  │  ┌────────────────┐ ┌──────────┐ ┌─────────┐ │               │   │
│  │  │ 🔴 火控雷达     │ │📷 光电   │ │🚢 航海  │ │               │   │
│  │  │ :3001          │ │:3002     │ │:3003    │ │               │   │
│  │  │ scan 60Hz      │ │fps 30    │ │20kn→0°  │◄┘               │   │
│  │  │ range 150km    │ │λ=1550nm  │ │S'hai    │   启动拉取       │   │
│  │  │ 32 targets     │ │1280x720  │ │auto:off  │   5次重试        │   │
│  │  │ mode:search    │ │sens=5    │ │PortA,B,C │   降级启动       │   │
│  │  └────────────────┘ └──────────┘ └─────────┘                   │   │
│  │   + 原有 default (order/user service, :3000)               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  调用链路:                                                            │
│  前端 ──(读/写)──► 后端 ──(读写)──► ConfigStore (内存)               │
│  多域服务 ──(只读)──► 后端 ──(读)──► ConfigStore.PublishedData        │
│  服务启动 ──(拉取)──► PullConfig() ──► ParseConfig() ──► 诊断        │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 三、项目文件结构

```
fse/
├── backend/                        # Go + Gin 配置中心后端
│   ├── main.go                     # 入口: seed data → CORS → 8 路由注册 → :8080
│   ├── handler/
│   │   ├── config.go               # 8 个 REST API handler
│   │   └── handler_test.go         # 28 个 API 集成测试 (含并发测试)
│   ├── store/
│   │   ├── store.go                # 数据模型 + 内存存储 (sync.RWMutex)
│   │   └── store_test.go           # 8 个单元测试
│   ├── go.mod / go.sum
│
├── frontend/                       # Vue 3 + Element Plus 管理界面
│   ├── src/
│   │   ├── main.js                 # 入口: createApp → ElementPlus → mount
│   │   ├── App.vue                 # 根组件: Dashboard + 3 Tab + 对话框
│   │   ├── style.css               # 全局设计 token (Linear 风格)
│   │   ├── api/
│   │   │   └── configService.js    # Axios API 封装 (5 个接口函数)
│   │   ├── utils/
│   │   │   └── configDiff.js       # deepEqual / hasDraft / countDiff
│   │   └── components/
│   │       ├── Sidebar.vue         # 左侧边栏: 环境/状态过滤
│   │       ├── TopBar.vue          # 顶层栏: 搜索框
│   │       ├── ConfigTable.vue     # 配置列表表格
│   │       ├── ConfigDrawer.vue    # 配置详情面板 (desktop inline / mobile drawer)
│   │       ├── ChangeLog.vue       # 变更记录时间线
│   │       ├── StatusBadge.vue     # 状态徽章 (已发布/待发布)
│   │       ├── ServiceStatus.vue   # 微服务状态页 (Tab 2)
│   │       └── EnvironmentPanel.vue # 环境与推送页 (Tab 3)
│   ├── vite.config.js              # Vite 配置 (+ /api /demo proxy)
│   ├── package.json
│
├── demo-service/                   # 多域仿真微服务平台 (单 binary)
│   ├── main.go                     # 入口: SERVICE_TYPE 分发 → 领域服务
│   ├── pkg/
│   │   └── configclient/           # 共享 ConfigClient SDK
│   │       ├── client.go           #   PullConfig / GetString / GetInt / GetFloat / GetEnum
│   │       └── diagnostics.go      #   通用配置诊断
│   ├── internal/
│   │   ├── radar/                  # 火控雷达服务
│   │   │   ├── config.go           #   RadarConfig + Parse
│   │   │   ├── target.go           #   Target 模型 + 状态机 + 生成器
│   │   │   ├── service.go          #   后台 goroutine (目标生成/更新)
│   │   │   └── handler.go          #   /radar/status|targets|mode
│   │   ├── sensor/                 # 光电传感器服务
│   │   │   ├── config.go           #   SensorConfig + Parse
│   │   │   ├── frame.go            #   Frame/Detection 模型 + 帧生成器
│   │   │   ├── service.go          #   后台 goroutine (帧生成, ≤30fps)
│   │   │   └── handler.go          #   /sensor/status|frame|calibrate
│   │   └── navigation/             # 船舶航海服务
│   │       ├── config.go           #   NavConfig + Parse
│   │       ├── position.go         #   GeoPosition/Waypoint + 位置计算
│   │       ├── service.go          #   后台 goroutine (GPS 位置更新)
│   │       └── handler.go          #   /nav/status|position|route|steer
│   ├── go.mod / go.sum
│
├── tests/
│   └── smoke_test.sh               # 全栈冒烟测试 (17 检查点 × 6 阶段)
│
├── start.sh                        # 一键启动: 后端 → 微服务 → 前端
├── ARCHITECTURE.md                 # 本文档
├── README.md                       # 安装部署文档
├── AI_PROMPTS.md                   # AI 协同研发报告
├── BUGLOG.md                       # Bug 追踪 (BUG-001 ~ BUG-012)
└── LINTLOG.md                      # 代码质量记录
```

---

## 四、数据模型

### 4.1 ConfigGroup (配置组)

```go
type ConfigGroup struct {
    Service          string            `json:"service"`          // 服务名, e.g. "order-service"
    Env              string            `json:"env"`              // 环境, e.g. "dev"/"test"/"prod"
    PublishedVersion string            `json:"publishedVersion"` // semver "X.Y.Z", e.g. "1.0.0"
    PublishedData    map[string]string `json:"publishedData"`    // 线上生效的配置
    DraftData        map[string]string `json:"draftData"`        // 编辑中、未发布的配置
    LastPublishedAt  time.Time         `json:"lastPublishedAt"`  // 最后发布时间
}
```

**隔离键**: `service:env` (如 `order-service:dev`)

**状态判定** (无 Status 字段，数据驱动):
- `reflect.DeepEqual(DraftData, PublishedData)` → 已发布 (published)
- `!DeepEqual` → 待发布 (pending)

**版本号**: `BumpVersion()` 递增补丁号: `""` → `"1.0.0"` → `"1.0.1"` → ...

### 4.2 ChangeRecord (变更记录)

```go
type ChangeRecord struct {
    Time     time.Time `json:"time"`     // 操作时间
    Action   string    `json:"action"`   // "新增" / "修改" / "删除" / "发布"
    Key      string    `json:"key"`      // 配置项 Key
    Version  string    `json:"version"`  // 当时版本号
    Operator string    `json:"operator"` // 操作人 (MVP 固定 "admin")
}
```

每组 `service:env` 最多保留 **10 条**记录。

### 4.3 ConfigStore (存储层)

```go
type ConfigStore struct {
    sync.RWMutex                                  // 读写锁
    Configs map[string]*ConfigGroup               // key: "service:env"
    Logs    map[string][]ChangeRecord             // key: "service:env"
}
```

---

## 五、后端 API 设计

基础地址: `http://localhost:8080`

| # | Method | Path | 说明 | 读/写锁 |
|---|--------|------|------|---------|
| 1 | `GET` | `/api/configs` | 列出所有配置组 | RLock |
| 2 | `GET` | `/api/configs/:service/:env` | 获取单个配置组 + 变更日志 | RLock |
| 3 | `GET` | `/api/configs/:service/:env/published` | 获取线上配置 (微服务拉取) | RLock |
| 4 | `PUT` | `/api/configs/:service/:env/keys/:key` | 新增/修改配置项 → DraftData | Lock |
| 5 | `DELETE` | `/api/configs/:service/:env/keys/:key` | 删除配置项 → DraftData | Lock |
| 6 | `POST` | `/api/configs/:service/:env/publish` | 发布: DraftData → PublishedData | Lock |
| 7 | `POST` | `/api/configs/:service/:env/push` | 推送 (预留) | RLock |
| 8 | `GET` | `/api/configs/watch` | SSE 监听 (预留) | RLock |

### SetKey 幂等性

若 key 不存在则创建 (action="新增")；若 value 相同则跳过日志 (action="修改" 仅记录变化)。Key 支持 URL 编码 (中文、`@` 等特殊字符通过 `encodeURIComponent` 处理)。

### Publish 语义

`CloneMap(DraftData)` 深拷贝到 `PublishedData`，版本号递增补丁号，记录 `LastPublishedAt`。

### CORS

允许 `localhost:5173` / `127.0.0.1:5173`，支持 GET/POST/PUT/DELETE/OPTIONS。

---

## 六、前端架构

### 6.1 页面结构

```
App.vue
│
├── Dashboard (可折叠)
│   ├── 4 统计卡片: 配置组 | 已发布 | 待发布 | 环境数
│   ├── 待发布队列 (clickable chips)
│   └── 操作栏: 时间戳 + 刷新 + 查看待发布 + 帮助(?)
│
└── Workspace (Tab 容器)
    ├── Tab Header: 配置管理 | 微服务状态 | 环境与推送 [+ 新建配置组]
    │
    ├── [Tab 1: 配置管理]
    │   ├── Sidebar (环境/状态过滤 + 服务列表)
    │   ├── ConfigMain
    │   │   ├── panel-head (标题 + 描述)
    │   │   ├── TopBar (搜索框)
    │   │   └── ConfigTable (配置列表表格)
    │   └── ConfigDrawer (inline panel / mobile drawer)
    │       ├── Header (服务名/环境/版本/状态)
    │       ├── 新增配置项 (独立表单卡片)
    │       ├── Diff 表格 (键 | 已发布 | 草稿 | 操作)
    │       └── ChangeLog + 发布按钮
    │
    ├── [Tab 2: 微服务状态]
    │   └── ServiceStatus (demo-service 检测 + 配置绑定展示)
    │
    └── [Tab 3: 环境与推送]
        └── EnvironmentPanel (环境隔离卡片 + 预留接口说明)
```

### 6.2 组件树

```
App.vue
├── TopBar.vue              # 搜索框
├── Sidebar.vue             # 左侧过滤: 全部/dev/test/prod + 已发布/待发布
├── ConfigTable.vue         # 配置列表: 服务/环境 | 版本 | 预览 | 变更 | 时间 | 状态
├── ConfigDrawer.vue        # 详情面板: Diff 对比 + 编辑 + 新增 + 删除 + 发布
│   ├── StatusBadge.vue     # 状态徽章
│   └── ChangeLog.vue       # 变更记录时间线
├── ServiceStatus.vue       # 微服务状态检测页
└── EnvironmentPanel.vue    # 环境隔离与推送预留页
```

### 6.3 数据流

```
getAllConfigs() ──► configs (ref) ──► filteredConfigs (computed)
                                         │
                    ┌────────────────────┼────────────────────┐
                    ▼                    ▼                    ▼
              ConfigTable          Sidebar (counts)     Dashboard (stats)
                    │
              openDetail(row) ──► selectedService / selectedEnv
                    │
                    ▼
              getOneConfig() ──► ConfigDrawer (diff table)
                    │
              PUT/DELETE/POST ──► fetchData() ──► 全量刷新
```

**状态管理**: 无 Vuex/Pinia，通过 `ref` + `computed` + Props/Emit 完成组件间通信。

### 6.4 API 封装 (`configService.js`)

| 函数 | HTTP | 用途 |
|------|------|------|
| `getAllConfigs()` | GET /api/configs | 初始加载 + 5s 轮询 |
| `getOneConfig(service, env)` | GET /api/configs/:svc/:env | 加载详情 |
| `setKey(service, env, key, value)` | PUT /keys/:key | 新增/修改 DraftData |
| `deleteKey(service, env, key)` | DELETE /keys/:key | 删除 DraftData 项 |
| `publishConfig(service, env)` | POST /publish | 发布 |

### 6.5 响应式设计

| 视口 | 布局 |
|------|------|
| ≥ 1100px | Sidebar (左 220px) + Table (中间) + ConfigDrawer inline panel (右, min(680px, 44vw)) |
| < 1100px | 无 Sidebar + Table 全宽 + ConfigDrawer 作为 el-drawer 弹出 |

### 6.6 设计 Token (Linear 风格)

```css
--bg-app: #f7f8fa;          /* 页面底色 */
--bg-surface: #ffffff;      /* 卡片/面板背景 */
--bg-selected: #eef0f9;     /* 选中态 */
--color-primary: #5e6ad2;   /* 主色调 (淡紫蓝) */
--color-published: #059669; /* 已发布 (绿) */
--color-pending: #d97706;   /* 待发布 (琥珀) */
--font-family: 'Inter', ...;/* 字体 */
--radius-lg: 8px;           /* 卡片圆角 */
--shadow-sm: 0 1px 2px rgba(0,0,0,0.04); /* 极轻阴影 */
```

---

## 七、核心流程

### 7.1 微服务启动拉取配置 (通用流程)

```
[demo-service 启动, SERVICE_TYPE=radar|sensor|navigation|default]
  │
  ├── 读取环境变量:
  │     SERVICE_TYPE=radar              — 选择运行哪个领域服务
  │     CONFIG_SERVICE=radar-service    — 服务标识
  │     CONFIG_ENV=dev                  — 环境
  │     CONFIG_URL=http://localhost:8080 — 配置中心地址
  │     RADAR_PORT=3001                 — (可选的) 端口覆盖
  │
  ├── NewConfigClient(configURL, service, env)
  │
  ├── PullConfig() [重试: 5次, 间隔 1s]
  │     └── GET /api/configs/{service}/{env}/published
  │           ├── 成功 → rawConfig map
  │           └── 失败 → 使用默认配置 (降级启动)
  │
  ├── ParseXxxConfig(client) — 类型安全解析 (以雷达为例):
  │     cfg.ScanRateHz = client.GetInt("radar.scan_rate_hz", 60, 1, 300)
  │     cfg.Mode       = client.GetEnum("radar.mode", "search", ["search","track","engage"])
  │     cfg.Band       = client.GetEnum("radar.band", "X", ["X","S","C","Ku"])
  │     ...
  │     字段缺失 → 使用默认值 | 字段值非法 → 警告 + 默认值
  │
  ├── PrintDiagnostics() — 启动诊断: raw keys / 未知 key 警告 / 跨服务配置检查
  │
  ├── 创建领域服务 + Start(ctx) → 启动后台 goroutine
  │
  └── r.Run(":" + port)
        ├── 领域 API (如 /radar/status, /sensor/frame, /nav/position)
        └── GET /health → {"status":"ok|degraded","config_from_center":true|false}
```

### 7.2 多域服务概览

demo-service 是一个 **单一 Go binary**，通过 `SERVICE_TYPE` 环境变量切换运行模式。每个领域服务有自己的**配置契约、业务状态机和领域 API**。

| 服务 | TYPE | 端口 | 配置数 | 领域 API | 后台任务 |
|------|------|------|--------|----------|----------|
| 🔴 火控雷达 | `radar` | 3001 | 5 keys | /radar/status, /radar/targets, /radar/mode | 目标生成 (scan_rate_hz) + 位置更新 (100ms) |
| 📷 光电传感器 | `sensor` | 3002 | 5 keys | /sensor/status, /sensor/frame, /sensor/calibrate | 帧生成 (≤30fps) |
| 🚢 船舶航海 | `navigation` | 3003 | 6 keys | /nav/status, /nav/position, /nav/route, /nav/steer | GPS 位置更新 (update_interval_s) |
| 📦 默认 | `default` | 3000 | 3 keys | GET / (配置回显) | 无 |

#### 火控雷达配置契约

| 配置 Key | 默认值 | 校验 | 说明 |
|----------|--------|------|------|
| radar.scan_rate_hz | 60 | 1-300 | 扫描频率 (Hz) |
| radar.range_km | 150 | 1-500 | 探测距离 (km) |
| radar.max_targets | 32 | 1-256 | 最大跟踪目标数 |
| radar.mode | "search" | search/track/engage | 工作模式 |
| radar.band | "X" | X/S/C/Ku | 雷达波段 |

- 目标状态机: `detected → tracking → locked → engaged` (进阶受 mode 影响)
- 目标类型: aircraft / missile / drone / ship
- 目标超出 range_km 自动从存储中移除

#### 光电传感器配置契约

| 配置 Key | 默认值 | 校验 | 说明 |
|----------|--------|------|------|
| sensor.wavelength_nm | 1550 | 400-14000 | 红外波长 |
| sensor.fps | 30 | 1-240 | 帧率 (实际限 30fps) |
| sensor.resolution | "1280x720" | 四种枚举 | 分辨率 |
| sensor.sensitivity | 5 | 1-10 | 灵敏度 |
| sensor.mode | "day" | day/night/thermal | 工作模式 |

- 检测概率 = f(mode, sensitivity): night=0.4x, thermal=1.5x, sensitivity 线性缩放
- 检测类型: person / vehicle / animal / aircraft (带 bbox + confidence)

#### 船舶航海配置契约

| 配置 Key | 默认值 | 校验 | 说明 |
|----------|--------|------|------|
| nav.update_interval_s | 5 | 1-300 | GPS 更新间隔 |
| nav.speed_knots | 20 | 1-60 | 巡航速度 (节) |
| nav.heading_deg | 0 | 0-359 | 航向 |
| nav.destination | "Shanghai" | — | 目的地 |
| nav.auto_pilot | "false" | true/false | 自动导航 |
| nav.waypoints | "PortA,PortB,PortC" | — | 航点列表 |

- 初始位置: 上海 (31.23N, 121.47E)
- auto_pilot=true 时自动计算航向指向下一航点/目的地
- 航点坐标是 hash-based 确定性的 (同名同坐标)
- auto_pilot 模式下不可手动 steer (返回 409)

### 7.3 种子配置 (10 组)

| # | service:env | 配置数 | 说明 |
|---|-------------|--------|------|
| 1-2 | order-service:dev/test | 3 | 原有 |
| 3-4 | user-service:dev/prod | 3 | 原有 |
| 5-6 | radar-service:dev/prod | 5 | dev=60Hz/search/X, prod=120Hz/track/C |
| 7-8 | sensor-service:dev/prod | 5 | dev=30fps/day/1550nm, prod=60fps/thermal/8000nm |
| 9-10 | navigation-service:dev/prod | 6 | dev=20kn/auto-off, prod=35kn/auto-on/Singapore |

### 7.4 前端编辑配置并发布

```
[用户操作]
  │
  ├── Tab 1: 配置管理 → 左侧 Sidebar 选环境/状态 → 点击表格行
  │     └── GET /api/configs/:service/:env → 右侧显示 Diff 对比面板
  │
  ├── 编辑 Draft 值
  │     ├── 点击 Draft 列值 → 组件级 editingKey 激活输入框
  │     ├── 修改值 → 回车
  │     └── PUT /api/configs/:service/:env/keys/:key → 更新 DraftData
  │
  ├── 新增配置项
  │     ├── 填写 Key + Value → 回车
  │     └── PUT /api/configs/:service/:env/keys/:newKey
  │
  ├── 删除配置项
  │     ├── 点击删除图标 → 确认
  │     └── DELETE /api/configs/:service/:env/keys/:key
  │
  └── 发布
        └── POST /api/configs/:service/:env/publish
              ├── CloneMap(DraftData) → PublishedData
              ├── BumpVersion() 递增补丁号
              └── 前端 5s 轮询刷新
```

### 7.5 前端自动刷新

```
setInterval(5000):
  GET /api/configs → 更新 configs(ref) → 所有 computed 自动重算 → UI 更新
```

---

## 八、测试覆盖

| 层级 | 工具 | 用例数 | 状态 |
|------|------|--------|------|
| 单元测试 (store) | go test | 8 | ✅ PASS |
| 集成测试 (handler) | go test + httptest | 28 | ✅ PASS |
| 冒烟测试 | bash | 17 (6阶段) | ✅ PASS |
| 前端构建 | npm run build | — | ✅ PASS |
| Lint | revive+go vet+gofmt | 13 issues | ✅ CLEAN |

---

## 九、扩展性设计

| 当前 MVP | 可扩展方向 |
|----------|-----------|
| `POST /push` 预留空实现 | SSE / WebSocket / 长轮询配置动态推送 |
| `GET /watch` 预留 SSE 端点 | 客户端订阅配置变更事件 |
| `BumpVersion()` 递增补丁号 | 配置回滚 + 历史版本查询 + 版本 Diff |
| `ConfigStore` 内存存储 | MySQL / SQLite / etcd 持久化 |
| `service:env` 二级隔离 | Namespace + Group + DataId 多级隔离 |
| `ChangeLog` 内存队列 (10条) | 数据库审计表 + 分页查询 |
| 单机部署 | 集群部署 + Raft/Paxos 一致性同步 |
| `ConfigClient.pullConfig()` | SDK 封装 + 本地文件缓存 + Watch 热更新回调 |
| 单 domain (default) | 多域仿真 (radar/sensor/navigation) + SERVICE_TYPE 分发 |
| 硬编码 seed config | 配置组模板库 + 一键导入 |
| 前端无路由 | Vue Router 多页面导航 |
| 无认证 | JWT / OAuth2 接入控制 |

---

## 十、技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 后端框架 | Go + Gin | 1.22.5 + 1.9.1 |
| 前端框架 | Vue 3 (Composition API) | 3.5 |
| UI 组件库 | Element Plus | 2.14 |
| HTTP 客户端 | Axios | 1.17 |
| 构建工具 | Vite | 8.0 |
| 示例服务 | Go + Gin | 1.22.5 + 1.9.1 |
| 测试 | Go testing + httptest + Bash | — |
| 代码质量 | revive + go vet + gofmt | — |
