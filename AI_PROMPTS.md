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
