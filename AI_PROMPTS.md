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
| `backend/store/store.go` | 90% | 数据模型 + CRUD 方法 + 线程安全 |
| `backend/handler/config.go` | 95% | 8 个 REST API handler |
| `backend/main.go` | 100% | CORS + 路由 + seed data |
| `backend/store/store_test.go` | 100% | 7 个单元测试 |
| `frontend/src/components/*` | 85% | 7 个 Vue 组件 |
| `frontend/src/api/configService.js` | 100% | API 封装层 |
| `demo-service/main.go` | 100% | ConfigClient SDK 雏形 |
| `start.sh` | 100% | 一键启动 + 进程管理 |
| `README.md` + `ARCHITECTURE.md` | 100% | 文档 |

## 遇到的 Bug 及修复

### Bug 1: Go 工具链版本不兼容
- **现象**: `go get` 下载了 go 1.25 工具链（因最新 gin/cors 要求），导致本地 go 1.22 无法构建
- **原因**: `go mod tidy` 自动升版本
- **修复**: 指定旧版本 `gin@v1.9.1 cors@v1.5.0`，并使用 `GOTOOLCHAIN=local` 强制本地工具链

### Bug 2: TopBar.vue 标签未闭合
- **现象**: `npm run build` 报 `Invalid end tag` 错误
- **原因**: `<button>` 标签后误写 `</span>` 而非 `</button>`
- **修复**: 改为 `</button>`

### Bug 3: ConfigTable.vue const 变量重赋值
- **现象**: `preview` 声明为 `const` 但随后 `preview += '...'`
- **修复**: 改为 `let`

### Bug 4: apt 只读文件系统
- **现象**: `apt-get install golang-go` 失败
- **原因**: 容器文件系统部分只读
- **修复**: 手动下载 Go 二进制包到 `/root/go`
