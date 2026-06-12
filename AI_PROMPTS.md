# AI 协同研发报告 (AI_PROMPTS.md)

## 使用的 AI 工具

| 工具 | 用途 |
|------|------|
| Claude Code (Reasonix) | 全流程：架构设计、数据模型、代码生成、测试、调试 |

## 关键 Prompt / Skill 记录

### 1. 设计阶段

**Prompt**: "请审查这个数据模型设计..."
- 触发技能的启用确认，一步步审查数据模型、API 设计、前端设计
- 输出: 19 条优化建议全部采纳

### 2. 前端设计

**Prompt**: "请你读取 linear.app 网站的风格再设计一下..."
- 使用 web_fetch 抓取 Linear.app 官网
- 提炼出设计 token：白色背景 #FAFAFA、极细 1px 边框、Inter 字体、4px 网格、克制色彩
- 输出: 12 条严格设计约束

### 3. Skill 使用

| Skill | 作用 |
|-------|------|
| `frontend-design` | 参考了其设计哲学（考究间距、字体、微动效），但适配到管理后台场景 |
| `explore` (未使用) | 可在代码审查时使用 |
| `review` (未使用) | 可在 PR 前使用 |

## AI 辅助完成的代码模块

| 模块 | AI 完成度 | 说明 |
|------|----------|------|
| `backend/store/store.go` | 90% | 数据模型 + CRUD 方法 + 线程安全 + BumpVersion() semver |
| `backend/handler/config.go` | 95% | 8 个 REST API handler |
| `backend/main.go` | 100% | CORS + 路由 + seed data |
| `backend/store/store_test.go` | 100% | 8 个单元测试 (含 BumpVersion) |
| `backend/handler/handler_test.go` | 100% | 28 个 API 集成测试 (含 6 子测试) |
| `frontend/src/components/*` | 85% | 7 个 Vue 组件 |
| `frontend/src/api/configService.js` | 100% | API 封装层 |
| `demo-service/main.go` | 100% | ConfigClient SDK 雏形 (含重试机制) |
| `tests/smoke_test.sh` | 100% | 全栈冒烟测试 (17 检查点) |
| `start.sh` | 100% | 一键启动 + 进程管理 + Go PATH 探测 |
| `README.md` + `ARCHITECTURE.md` | 100% | 文档 |
| `BUGLOG.md` + `LINTLOG.md` | 100% | Bug 追踪 + Lint 记录 |
| `backend/.golangci.yml` | 100% | golangci-lint 配置 (备用) |

## 遇到的 Bug 及修复

### BUG-001: Go 工具链版本不兼容
- **现象**: `go get` 下载了 go 1.25 工具链（因最新 gin/cors 要求），导致本地 go 1.22 无法构建
- **原因**: `go mod tidy` 自动升版本
- **修复**: 指定旧版本 `gin@v1.9.1 cors@v1.5.0`，并使用 `GOTOOLCHAIN=local` 强制本地工具链

### BUG-002: TopBar.vue 标签未闭合
- **现象**: `npm run build` 报 `Invalid end tag` 错误
- **原因**: `<button>` 标签后误写 `</span>` 而非 `</button>`
- **修复**: 改为 `</button>`

### BUG-003: ConfigTable.vue const 变量重赋值
- **现象**: `preview` 声明为 `const` 但随后 `preview += '...'`
- **修复**: 改为 `let`

### BUG-004: apt 只读文件系统
- **现象**: `apt-get install golang-go` 失败
- **原因**: 容器文件系统部分只读
- **修复**: 手动下载 Go 二进制包到 `/root/go`

### BUG-005: demo-service 后端未就绪时直接 Fatal 退出
- **现象**: 后端未启动时 `go run .` 直接 `log.Fatal` 退出
- **原因**: `pullConfig()` 无重试机制，一次失败即 `log.Fatalf`
- **修复**: `pullConfig()` 加重试循环 (5次, 间隔1s)，`main()` 中 `log.Fatalf` 降级为 `log.Printf` + 默认端口启动

### BUG-006: start.sh 找不到 Go
- **现象**: `bash start.sh` 报 `go: command not found`
- **原因**: WSL 中 Go 装在 `/root/go/bin`，不在默认 PATH 中
- **修复**: `start.sh` 开头增加 4 个常见 Go 路径自动探测 + `command -v` 校验 + 清晰报错

### BUG-007: ConfigDrawer 点击行空白
- **现象**: 点击表格任意一行 → 右侧抽屉/面板完全空白，无任何按钮
- **原因**: `<script>` (非 setup) 块定义的 `DetailBody` 子组件模板中使用 `<ChangeLog>`，但 `ChangeLog` 只在 `<script setup>` 中 import，两个 script 块作用域隔离
- **修复**: 删除非 setup 的 `<script>` 块，将 `DetailBody` 模板内联到 panel 和 drawer 中

## TDD 测试阶段

### 第一轮: 单元测试 (Store 层)

```bash
cd backend && go test ./store/ -v
```

| 测试 | 覆盖 |
|------|------|
| TestSetKeyDraftOnly | 编辑只改 DraftData，PublishedData 不变 |
| TestPublishVersionIncrease | 发布后版本号递增 ("1.0.0" → "1.0.1") |
| TestBumpVersion | BumpVersion 边界 (空串/正常/大号/非法) |
| TestPublishDeepCopy | DeepCopy 保护，DraftData 修改不污染 PublishedData |
| TestHasDraft | reflect.DeepEqual 正确判定 Draft/Published 差异 |
| TestChangeLogLimit | 变更记录保留最近 10 条 |
| TestCloneMap | CloneMap 深拷贝不泄露 |
| TestDeleteKey | 删除只影响目标 key |

**结果: 8/8 PASS**

### 第二轮: API 集成测试 (Handler 层)

```bash
cd backend && go test ./handler/ -v
```

| 测试 | 覆盖 |
|------|------|
| TestHandler_ListAll_Empty/WithData | GET /api/configs |
| TestHandler_GetOne_Existing/NonExisting | GET /api/configs/:service/:env |
| TestHandler_SetKey_Success/MissingBody/URLEncodedKey | PUT /keys/:key (正常/400/编码) — 含 6 子测试 |
| TestHandler_DeleteKey_Existing/NonExisting | DELETE /keys/:key (幂等) |
| TestHandler_Publish | POST /publish → Version + DeepCopy |
| TestHandler_GetPublished/NonExisting | GET /published |
| TestHandler_Push_Reserved | POST /push 预留 |
| TestHandler_Watch_Reserved | GET /watch 预留 |
| TestHandler_FullFlow_EditPublishVerify | 6 步全流程: SET→检查→PUBLISH→验证→EDIT→确认隔离 |
| TestHandler_ConcurrentWrites | 10 goroutine 并发写入 + Publish |

**结果: 28/28 PASS**

### 第三轮: 全栈冒烟测试

```bash
bash tests/smoke_test.sh
```

6 阶段 17 检查点:
1. 编译 backend + demo-service (2)
2. 启动后端 (1)
3. API 基础验证 (6)
4. 编辑 → 发布 完整流程 (5)
5. Demo Service 拉取配置 (2)
6. 结果汇总 (1)

**结果: 17/17 PASS**

## Lint 代码质量检查

| 工具 | 版本 | 覆盖率 | 发现 | 修复 | 当前 |
|------|------|--------|------|------|------|
| revive | v1.7.0 | 风格/命名/注释 | 12 issues | 12 | CLEAN |
| go vet | go1.22.5 | 语义/并发 | 0 | — | CLEAN |
| gofmt | go1.22.5 | 格式化 | 1 diff | 1 | 0 diffs |
| golangci-lint | v1.55.2 | 全量聚合 | ⚠️ (备用) | — | 已安装 |

详见 `LINTLOG.md`。

## 第四轮: 多域仿真平台扩展 (2026-06-12)

### 设计 Prompt

**用户**: "demo-service可以有业务逻辑实现吗，例如三个服务甚至更多 模拟航空中的火控雷达，光电中的传感器，船舶中的航海等等"

**AI 响应**: 先使用 Plan Agent 设计架构，再按 8 步顺序实施:
1. 提取 ConfigClient SDK (`pkg/configclient/`)
2. 火控雷达服务 (`internal/radar/`)
3. 光电传感器服务 (`internal/sensor/`)
4. 船舶航海服务 (`internal/navigation/`)
5. main.go 重写为 SERVICE_TYPE 分发器
6. 后端种子数据 4→10 组
7. start.sh 多进程
8. 文档更新

### 关键 Prompt 记录

| 阶段 | Prompt | 工具/技能 |
|------|--------|----------|
| 架构设计 | "设计一个详细的实现计划来扩展 demo-service" | Plan Agent |
| SDK 提取 | "提取 ConfigClient 到 pkg/configclient/" | Edit/Write |
| 雷达服务 | "实现火控雷达服务: 配置契约 + 状态机 + 后台 goroutine" | Write (4 files) |
| 传感器服务 | "实现光电传感器: 帧生成 + 检测模型 + 环形缓冲" | Write (4 files) |
| 航海服务 | "实现船舶航海: GPS 推进 + 航点追踪 + 自动导航" | Write (4 files) |
| main.go 重写 | "重写 main.go 为 SERVICE_TYPE 分发器 + CORS 中间件" | Write |
| 后端种子数据 | "backend/main.go 种子数据加 6 个新配置组" | Edit |
| start.sh 更新 | "start.sh 增加 :3000 默认服务" | Write |
| 概念澄清 | "配置组 vs 配置项的定义和 count 计算" | — |
| 微服务状态修复 | "ServiceStatus 滚轮不生效/配置已绑定判断有误/服务列表硬编码" | Write |
| 图标修复 | "刷新和帮助图标缺失" | Edit (改用 Element Plus el-icon) |
| 帮助文档重写 | "帮助文档阐述配置组和配置项的区别" | Edit |
| 评估报告优化 | "读取 fde-candidate-assessment-report 并优化，不加数据库" | Write (交互版HTML) |
| 测试更新 | "更新 test 文件并跑测试" | Edit/Write (3 test files) |

### AI 辅助模块 (本轮新增)

| 模块 | AI 完成度 | 说明 |
|------|----------|------|
| `pkg/configclient/client.go` | 95% | 从 main.go 提取 + GetString/Int/Float/Enum 泛化 |
| `pkg/configclient/diagnostics.go` | 100% | 通用诊断函数 |
| `pkg/configclient/client_test.go` | 100% | 9 个 SDK 测试 (含 mock HTTP server) |
| `internal/radar/*.go` | 90% | 4 文件: config/target/service/handler |
| `internal/sensor/*.go` | 90% | 4 文件: config/frame/service/handler |
| `internal/navigation/*.go` | 90% | 4 文件: config/position/service/handler |
| `demo-service/main.go` | 85% | SERVICE_TYPE 分发器 + CORS |
| `frontend/.../ServiceStatus.vue` | 95% | /health 真实探测 + 4 服务 + 滚轮修复 |
| `frontend/.../Sidebar.vue` | 95% | 服务列表聚合 + 过滤互斥 |
| `frontend/.../App.vue` | 90% | 图标修复 + 帮助重写 + 稳定排序 |
| `backend/store/store_test.go` | 100% | +4 测试 |
| `backend/handler/handler_test.go` | 100% | +3 测试 |
| `tests/smoke_test.sh` | 100% | +11 检查点 (领域服务) |
| `DEMO_SERVICES.md` | 100% | 领域服务完整说明 |
| `fde-candidate-assessment-report` | 95% | 交互版 HTML (折叠/深色/笔记/localStorage) |

### 本轮 Bug 及修复

#### BUG-013: 前端配置列表排序随机跳动
- **现象**: 每 5 秒轮询后表格行顺序随机变化
- **原因**: 后端 `ConfigStore.Configs` 是 Go `map`，遍历顺序随机，JSON 数组顺序每次不同
- **修复**: `filteredConfigs` 末尾加稳定排序 (service 字母序 → env 固定顺序 dev/test/prod)

#### BUG-014: 微服务状态页"配置已绑定"语义错误
- **现象**: user-service:dev 显示"配置已绑定"，但根本没有运行实例
- **原因**: 判断逻辑只检查后端有无对应配置组，不等同于服务在运行
- **修复**: 重写为 `/health` 端点真实探测，仅展示有 endpoint 的 4 个服务

#### BUG-015: 微服务状态页无法滚动
- **现象**: 服务卡片超出视口时鼠标滚轮无反应
- **原因**: `.service-status` 未正确约束高度，`overflow:auto` 不生效
- **修复**: `.service-scroll` 使用 `flex:1; overflow-y:auto` + 父容器 `height:100%`

#### BUG-016: CORS 缺失导致前端无法探测 /health
- **现象**: 前端 :5173 向 :3001/:3002/:3003 发 `/health` 请求被浏览器拦截
- **原因**: demo-service 未设置 CORS 头
- **修复**: main.go 添加 `corsMiddleware()` 设置 `Access-Control-Allow-Origin: *`

#### BUG-017: 刷新和帮助图标渲染异常
- **现象**: 自绘 SVG 弧线显示不完整
- **修复**: 改用 Element Plus `<el-icon><Refresh /></el-icon>` 和 `<QuestionFilled />`

## 测试覆盖更新

### 第四轮测试结果

| 层级 | 用例数 | 新增 | 结果 |
|------|--------|------|------|
| Store 单元测试 | 12 | +4 | ✅ PASS |
| Handler 集成测试 | 21 | +3 | ✅ PASS |
| ConfigClient SDK | 9 | +9 | ✅ PASS |
| 冒烟测试 | 22 | +11 | ✅ 22/22 PASS |
| 前端构建 | — | — | ✅ PASS |
<<<<<<< HEAD
=======

## 第五轮: 交付收尾 (2026-06-12)

### Prompt 记录

| Prompt | 工具 | 产出 |
|--------|------|------|
| "检查是否满足交付物要求" | Read + Bash | 逐条核对 4.1~4.4 |
| "同网段的可以访问该网址" | Edit | start.sh: `--host 0.0.0.0` + LAN IP 自动探测 |
| "更新文档并提交" | Write/Edit | README/ARCHITECTURE/AI_PROMPTS 终稿 |
| "整理 skills 和 prompt 文档" | Write | SKILLS_AND_PROMPTS.md |
| "git clone 之后直接 bash start.sh 就可以启动" | Edit/Bash | start.sh 加 `npm install` 自动检测 |

### 交付物查漏补缺

| 问题 | 修复 |
|------|------|
| `npm run dev` 默认只监听 127.0.0.1 | Vite 加 `--host 0.0.0.0`，同网段可访问 |
| ServiceStatus 写死 localhost | 改为 `window.location.hostname` 动态拼接 |
| `start.sh` 未检测 node_modules | 加 `if [ ! -d node_modules ]` 自动 `npm install` |
| 缺少 Skills 整理文档 | 新建 SKILLS_AND_PROMPTS.md |
>>>>>>> a77b208 (docs: 交付收尾 — 文档终稿 + start.sh 开箱即用 + 同网段访问)
