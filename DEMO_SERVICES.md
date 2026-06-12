# Demo Service — 多域仿真微服务平台

> 分布式配置中心的示例客户端，模拟 **3 个不同领域的真实业务系统**：火控雷达（航空）、光电传感器、船舶航海。单一 Go binary，通过 `SERVICE_TYPE` 环境变量切换。

---

## 一、架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                 demo-service (单 binary)                     │
│                                                             │
│  main.go  ── SERVICE_TYPE ──► ┌──────────┐                 │
│                               │ Dispatcher│                 │
│                               └────┬─────┘                 │
│               ┌───────────────────┼───────────────────┐    │
│               ▼                   ▼                   ▼    │
│  ┌─────────────────┐ ┌──────────────────┐ ┌──────────────┐ │
│  │ 🔴 火控雷达      │ │ 📷 光电传感器     │ │ 🚢 船舶航海   │ │
│  │ radar           │ │ sensor           │ │ navigation   │ │
│  │ :3001           │ │ :3002            │ │ :3003        │ │
│  │                 │ │                  │ │              │ │
│  │ 5 配置项        │ │ 5 配置项          │ │ 6 配置项      │ │
│  │ 目标状态机      │ │ 帧生成 + 检测     │ │ GPS + 航点    │ │
│  │ 2 后台 goroutine│ │ 1 后台 goroutine  │ │ 1 后台 groutine│ │
│  │ 4 HTTP 端点     │ │ 4 HTTP 端点       │ │ 4 HTTP 端点   │ │
│  └─────────────────┘ └──────────────────┘ └──────────────┘ │
│                                                             │
│  ┌─────────────────┐                                       │
│  │ 📦 默认 (default)│  ← 向后兼容，配置回显模式               │
│  │ :3000           │                                       │
│  │ 3 配置项 (通用)  │                                       │
│  └─────────────────┘                                       │
└─────────────────────────────────────────────────────────────┘
```

**启动流程**（所有服务统一）：

```
读取环境变量 → ConfigClient.PullConfig() (5次重试)
  → ParseXxxConfig() (类型安全 + 逐字段校验)
  → PrintDiagnostics() (启动诊断报告)
  → XxxService.Start(ctx) (启动后台协程)
  → 注册 HTTP 路由 → 开始接受请求
```

**配置中心不可达时**：打印警告，全部使用 `DefaultConfig()` 降级启动，`/health` 返回 `"degraded"`。

---

## 二、火控雷达 (radar-service) 🔴

> 模拟航空火控雷达的目标检测与跟踪系统。

### 配置契约

| 配置 Key | 类型 | 默认值 | 校验 | 说明 |
|----------|------|--------|------|------|
| `radar.scan_rate_hz` | int | 60 | 1–300 | 扫描频率，决定目标生成速率 |
| `radar.range_km` | int | 150 | 1–500 | 探测半径，超出此范围的目标被丢弃 |
| `radar.max_targets` | int | 32 | 1–256 | 同时跟踪的目标数上限 |
| `radar.mode` | enum | `"search"` | search/track/engage | 工作模式，影响目标状态机行为 |
| `radar.band` | enum | `"X"` | X/S/C/Ku | 雷达波段 (X=火控, S=中程, C=远程, Ku=短程) |

### 业务逻辑

**后台 Goroutine 1 — 目标生成器**：按 `scan_rate_hz` 频率在雷达范围内随机生成目标（类型: aircraft/missile/drone/ship），含三维位置和速度矢量。达到 `max_targets` 上限后暂停生成。

**后台 Goroutine 2 — 目标更新器**：每 100ms 遍历所有目标，更新位置（`pos += vel × dt`），推进状态机，移除超出 `range_km` 的目标。

**目标状态机**：

```
detected ──(~2s)──► tracking ──(2s)──► locked ──(2s)──► engaged
                                                    (仅 engage 模式)
 search 模式: 最多到 tracking
 track  模式: 最多到 locked
 engage 模式: 可进阶到 engaged
```

### API

| Method | Path | 说明 |
|--------|------|------|
| `GET` | `/radar/status` | 配置 + 运行时间 + 各状态目标数 (detected/tracking/locked/engaged) |
| `GET` | `/radar/targets` | 所有活跃目标列表（ID/类型/位置/速度/状态/时间） |
| `GET` | `/radar/targets/:id` | 单个目标详情，目标不存在返回 404 |
| `POST` | `/radar/mode` | `{"mode":"track"}` 动态切换工作模式，非法模式返回 400 |

### 示例响应

```json
// GET /radar/status
{
  "service": "radar-service",
  "config": {"scan_rate_hz":60, "range_km":150, "max_targets":32, "mode":"search", "band":"X"},
  "uptime_seconds": 15.4,
  "targets_detected": 120,
  "targets_tracked": 8,
  "targets_locked": 3,
  "targets_engaged": 0,
  "active_targets": 32
}
```

---

## 三、光电传感器 (sensor-service) 📷

> 模拟红外光电传感器的图像采集与目标检测系统。

### 配置契约

| 配置 Key | 类型 | 默认值 | 校验 | 说明 |
|----------|------|--------|------|------|
| `sensor.wavelength_nm` | int | 1550 | 400–14000 | 红外工作波长 (nm) |
| `sensor.fps` | int | 30 | 1–240 | 帧率，实际采集上限 30fps |
| `sensor.resolution` | enum | `"1280x720"` | 640x480/1280x720/1920x1080/3840x2160 | 图像分辨率 |
| `sensor.sensitivity` | int | 5 | 1–10 | 灵敏度，影响检测概率 |
| `sensor.mode` | enum | `"day"` | day/night/thermal | 工作模式，影响检测能力 |

### 业务逻辑

**后台 Goroutine — 帧生成器**：按 `fps` 频率生成模拟图像帧（实际限 30fps），存入 60 帧环形缓冲区。每帧包含随机检测结果。

**检测概率模型**：

| 因素 | 影响 |
|------|------|
| mode = `night` | 0.4× 检测概率 |
| mode = `thermal` | 1.5× 检测概率 |
| sensitivity = 1 | 0.3× 期望检测数 |
| sensitivity = 10 | 2.0× 期望检测数 |

检测结果包含：类型 (person/vehicle/animal/aircraft)、置信度 (0–1)、边界框 (像素坐标)。

### API

| Method | Path | 说明 |
|--------|------|------|
| `GET` | `/sensor/status` | 配置 + 运行时间 + 累计帧数/检测数 + 缓冲区大小 |
| `GET` | `/sensor/frame` | 最新一帧的完整数据 (含所有 detection + bbox) |
| `GET` | `/sensor/frames?count=N` | 最近 N 帧摘要（帧ID/时间戳/检测数），N≤60 |
| `POST` | `/sensor/calibrate` | 模拟校准，返回暗电流/增益/曝光/温度等参数 |

### 示例响应

```json
// GET /sensor/frame
{
  "frame": {
    "frame_id": 1350,
    "timestamp": "2026-06-12T18:26:15+08:00",
    "resolution": "1280x720",
    "mode": "day",
    "detections": [
      {"type":"person","confidence":0.82,"bbox":{"x":320,"y":180,"w":120,"h":240}},
      {"type":"vehicle","confidence":0.65,"bbox":{"x":600,"y":400,"w":80,"h":50}}
    ]
  }
}
```

---

## 四、船舶航海 (navigation-service) 🚢

> 模拟船舶 GPS 导航与航线追踪系统。

### 配置契约

| 配置 Key | 类型 | 默认值 | 校验 | 说明 |
|----------|------|--------|------|------|
| `nav.update_interval_s` | int | 5 | 1–300 | GPS 位置更新间隔 (秒) |
| `nav.speed_knots` | int | 20 | 1–60 | 巡航速度 (节) |
| `nav.heading_deg` | int | 0 | 0–359 | 航向角 (0=北, 90=东) |
| `nav.destination` | string | `"Shanghai"` | — | 最终目的地 |
| `nav.auto_pilot` | bool | `false` | true/false | 自动导航：自动计算航向指向下一目标 |
| `nav.waypoints` | csv | `"PortA,PortB,PortC"` | — | 逗号分隔的航点列表 |

### 业务逻辑

**后台 Goroutine — 位置更新器**：每 `update_interval_s` 秒推进 GPS 位置。

**航行模型**：
- 初始位置：上海 (31.23°N, 121.47°E)
- 速度转换：1 节 = 1 海里/小时 = 1852 m/h
- 使用简化球面几何计算经纬度增量
- 航点坐标：hash-based 确定性映射（同名航点永远在同坐标）
- 距航点 <10km 视为到达，自动推进到下一航点

**自动导航 vs 手动操控**：

| 模式 | 行为 | steer 接口 |
|------|------|-----------|
| `auto_pilot=false` | 按 `nav.heading_deg` 直航 | ✅ 可手动转向 |
| `auto_pilot=true` | 自动计算航向指向下一航点/目的地 | ❌ 返回 409 Conflict |

### API

| Method | Path | 说明 |
|--------|------|------|
| `GET` | `/nav/status` | 配置 + 当前位置 + 航向 + 速度 + ETA + 航点进度 |
| `GET` | `/nav/position` | 详细 GPS 位置 (经纬度/精度/时间戳) |
| `GET` | `/nav/route` | 完整航线：航点列表 + 完成进度 + 目的地坐标 + ETA |
| `POST` | `/nav/steer?heading=90` | 手动转向 (仅 auto_pilot=false 时可用) |

### 示例响应

```json
// GET /nav/status
{
  "service": "navigation-service",
  "current_position": {"lat":31.24, "lng":121.48, "timestamp":"...", "accuracy_m":3.5},
  "heading_deg": 0,
  "speed_knots": 20,
  "destination": "Shanghai",
  "eta": "83h 12m",
  "auto_pilot": false,
  "current_waypoint": "PortA",
  "waypoint_progress": "0/3"
}
```

---

## 五、默认模式 (default)

向后兼容的配置回显模式，保留原有的 `GET /` 和 `GET /health` 端点。

| 配置 Key | 默认值 | 校验 |
|----------|--------|------|
| `server.port` | 3000 | 1–65535 |
| `db.url` | `localhost:3306` | 任意非空 |
| `log.level` | `info` | debug/info/warn/error |

```
GET / → {"service","env","config":{...},"raw":{...}}
```

---

## 六、共享 SDK

`pkg/configclient/` 为所有领域服务提供统一的配置拉取和校验能力：

```go
// 创建客户端
client := configclient.New("http://localhost:8080", "radar-service", "dev")

// 拉取配置 (5次重试, 1s间隔)
client.PullConfig()

// 类型安全获取
client.GetString("radar.mode", "search")
client.GetInt("radar.scan_rate_hz", 60, 1, 300)
client.GetFloat("radar.gain", 1.0, 0.1, 10.0)
client.GetEnum("radar.band", "X", []string{"X","S","C","Ku"})

// 通用诊断
configclient.PrintDiagnostics("radar-service", client.GetRawConfig(), KnownKeys())
```

所有 Getter 校验失败时**打印警告 + 返回默认值**，保证降级启动不崩溃。

---

## 七、启动

### 一键启动全部

```bash
./start.sh
```

### 分别启动

```bash
# 后端 (必须先启动)
cd backend && go run .

# 火控雷达 (dev 环境, 5个配置项)
cd demo-service && \
  SERVICE_TYPE=radar CONFIG_SERVICE=radar-service CONFIG_ENV=dev RADAR_PORT=3001 go run .

# 光电传感器 (dev 环境, 5个配置项)
cd demo-service && \
  SERVICE_TYPE=sensor CONFIG_SERVICE=sensor-service CONFIG_ENV=dev SENSOR_PORT=3002 go run .

# 船舶航海 (dev 环境, 6个配置项)
cd demo-service && \
  SERVICE_TYPE=navigation CONFIG_SERVICE=navigation-service CONFIG_ENV=dev NAV_PORT=3003 go run .

# 默认模式
cd demo-service && go run .  # :3000, SERVICE_TYPE=default
```

### 环境变量对照

| 变量 | 用途 | 默认值 |
|------|------|--------|
| `SERVICE_TYPE` | 选择领域服务 | `default` |
| `CONFIG_SERVICE` | 配置中心中的服务标识 | `order-service` |
| `CONFIG_ENV` | 配置环境 | `dev` |
| `CONFIG_URL` | 配置中心地址 | `http://localhost:8080` |
| `RADAR_PORT` | 雷达服务监听端口 | `3001` |
| `SENSOR_PORT` | 传感器服务监听端口 | `3002` |
| `NAV_PORT` | 航海服务监听端口 | `3003` |
| `PORT` | 通用端口覆盖（优先级低于专用） | `3000` |

---

## 八、配置中心联动

配置中心后端 (`backend/main.go`) 预置了 10 个配置组：

| # | service:env | 领域 | 配置数 | 示例 |
|---|-------------|------|--------|------|
| 1 | order-service:dev | 通用 | 3 | db.url/server.port/log.level |
| 2 | order-service:test | 通用 | 3 | 同上 |
| 3 | user-service:dev | 通用 | 3 | 同上 |
| 4 | user-service:prod | 通用 | 3 | 同上 |
| 5 | **radar-service:dev** | 🔴 雷达 | 5 | 60Hz/search/X/150km/32tgt |
| 6 | **radar-service:prod** | 🔴 雷达 | 5 | 120Hz/track/C/300km/128tgt |
| 7 | **sensor-service:dev** | 📷 传感器 | 5 | 30fps/day/1280x720/sens=5 |
| 8 | **sensor-service:prod** | 📷 传感器 | 5 | 60fps/thermal/1920x1080/sens=8 |
| 9 | **navigation-service:dev** | 🚢 航海 | 6 | 20kn/auto-off/Shanghai |
| 10 | **navigation-service:prod** | 🚢 航海 | 6 | 35kn/auto-on/Singapore |

**dev vs prod 对比**展示了配置中心的核心价值：同一份代码、同一套业务逻辑，通过不同配置驱动完全不同的运行参数。

在配置中心前端修改配置 → 发布 → 重启领域服务 → 验证行为变化（如雷达从 60Hz/search 变为 120Hz/track）。

---

## 九、文件结构

```
demo-service/
├── main.go                        # 入口: SERVICE_TYPE 分发 + 端口确定 + 优雅关闭
│
├── pkg/configclient/              # 共享 SDK
│   ├── client.go                  #   ConfigClient + GetString/GetInt/GetFloat/GetEnum
│   └── diagnostics.go             #   PrintDiagnostics (通用配置诊断)
│
├── internal/
│   ├── radar/                     # 🔴 火控雷达
│   │   ├── config.go              #   配置契约 + 解析
│   │   ├── target.go              #   目标模型 + 状态机 + 生成器 + 存储
│   │   ├── service.go             #   RadarService: Start/Stop + 后台协程
│   │   └── handler.go             #   Gin 路由注册 + 4 个 Handler
│   │
│   ├── sensor/                    # 📷 光电传感器
│   │   ├── config.go              #   配置契约 + 解析
│   │   ├── frame.go               #   帧模型 + 检测模型 + 帧生成器 + 环形缓冲
│   │   ├── service.go             #   SensorService: Start/Stop + 帧生成协程
│   │   └── handler.go             #   Gin 路由注册 + 4 个 Handler
│   │
│   └── navigation/                # 🚢 船舶航海
│       ├── config.go              #   配置契约 + 解析 (含航点 CSV 切分)
│       ├── position.go            #   GPS/航点模型 + 位置推进 + 航向/ETA 计算
│       ├── service.go             #   NavService: Start/Stop + 位置更新协程
│       └── handler.go             #   Gin 路由注册 + 4 个 Handler
│
├── go.mod
└── go.sum
```
