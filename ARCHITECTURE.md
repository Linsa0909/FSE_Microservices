# 架构设计文档 (ARCHITECTURE.md)

## 一、整体架构拓扑

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                   │
│   ┌──────────────┐    轮询 (5s)    ┌──────────────────┐          │
│   │  前端 (Vue)  │ ◄────────────── │  配置中心后端     │          │
│   │  :5173       │ ──────────────► │  Go + Gin :8080  │          │
│   │              │   PUT/POST/DEL   │                  │          │
│   └──────────────┘                  │  ┌────────────┐  │          │
│                                     │  │ ConfigStore │  │          │
│   ┌──────────────┐                  │  │ (内存存储)  │  │          │
│   │ 示例微服务    │    拉取配置      │  └────────────┘  │          │
│   │ Go :3000     │ ◄────────────── │                  │          │
│   │              │   GET /published │                  │          │
│   └──────────────┘                  └──────────────────┘          │
│                                                                   │
│   调用链路:                                                        │
│   前端 ←→ 后端 ← 内存存储                                          │
│   微服务 → 后端 (启动时拉取 PublishedData)                          │
└──────────────────────────────────────────────────────────────────┘
```

## 二、核心流程

### 2.1 微服务启动拉取配置

```
[微服务启动]
  │
  ├── ConfigClient.pullConfig()
  │     │
  │     └── GET /api/configs/order-service/dev/published
  │           │
  │           └── 后端返回 PublishedData (JSON)
  │                 │
  │                 └── 客户端缓存到内存
  │                       │
  │                       └── 业务代码通过 GetConfig() 读取
  │
  └── 启动 HTTP 服务 (端口从配置 server.port 读取)
```

### 2.2 前端编辑配置并发布

```
[前端操作]
  │
  ├── 点击行 → 打开详情面板 (Drawer)
  │     │
  │     └── 显示 PublishedData vs DraftData 双列对比
  │
  ├── 编辑 Draft Value (内联编辑)
  │     │
  │     └── PUT /api/configs/:service/:env/keys/:key
  │           │
  │           └── 后端修改 DraftData, 记录 ChangeLog
  │
  ├── 新增配置项
  │     │
  │     └── PUT /api/configs/:service/:env/keys/:newKey
  │
  ├── 删除配置项
  │     │
  │     └── DELETE /api/configs/:service/:env/keys/:key
  │
  └── 点击 [发布]
        │
        └── POST /api/configs/:service/:env/publish
              │
              └── 后端: CloneMap(DraftData) → PublishedData
                        Version++, LastPublishedAt = now()
              │
              └── 前端 5s 轮询刷新列表
```

### 2.3 前端 5s 轮询

```
setInterval(5000):
  GET /api/configs
    → 更新表格数据
    → 更新"上次刷新"时间
```

## 三、数据模型

### ConfigGroup

```go
type ConfigGroup struct {
    Service          string            // 服务名
    Env              string            // 环境 (dev/test/prod)
    PublishedVersion string            // 已发布版本号, e.g. "1.0.0" (semver)
    PublishedData    map[string]string // 线上生效的配置
    DraftData        map[string]string // 编辑中的配置
    LastPublishedAt  time.Time         // 最后发布时间
}
```

**版本号规则**：采用 `vX.Y.Z` 三位语义化版本格式。首次发布为 `"1.0.0"`，每次发布时通过 `BumpVersion()` 递增补丁号 (`"1.0.0"` → `"1.0.1"` → `"1.0.2"`)。大版本 (`X`) 和中版本 (`Y`) 由主流程或人工调整。

**状态判定**（无 Status 字段，由数据驱动）：
- `DraftData == PublishedData` → PUBLISHED
- `DraftData != PublishedData` → PENDING_PUBLISH

### ChangeRecord

```go
type ChangeRecord struct {
    Time     time.Time // 操作时间
    Action   string    // 新增/修改/删除/发布
    Key      string    // 配置项 Key
    Version  string    // 当时版本号, e.g. "1.0.2"
    Operator string    // 操作人 (MVP 固定 "admin")
}
```

## 四、扩展性设计

| 当前 MVP | 可扩展为 |
|----------|---------|
| POST /push 预留空实现 | SSE / WebSocket / 长轮询推送 |
| Version 自增计数器 | 配置回滚 + 历史版本查询 |
| 内存存储 | MySQL / SQLite 持久化 |
| service:env 二级隔离 | Namespace + Group 多级隔离 |
| ChangeLog 内存队列 (10条) | 数据库审计表 |
| 单机部署 | 集群部署 + Raft 一致性同步 |
| ConfigClient pullConfig | SDK 封装 + 本地缓存 + 热更新回调 |
