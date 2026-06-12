# 🐛 Bug Tracking Log

> 分布式配置中心 — 开发阶段 Bug 追踪
> 格式: | ID | 状态 | 模块 | 发现时间 | 现象 | 根因 | 修复方案 | 验证方式 | 关联 Commit |

---

## 统计

| 状态 | 数量 |
|------|------|
| ✅ Closed | 4 |
| 🔧 Fixing | 0 |
| 👀 Known | 0 |

---

## 详细记录

| ID | 状态 | 模块 | 发现 | 现象 | 根因 | 修复 | 验证 | Commit |
|----|------|------|------|------|------|------|------|--------|
| BUG-001 | ✅ Closed | `backend/build` | 2026-06-12 | `go get` 下载 go 1.25 工具链，本地 go 1.22 无法构建 | `go mod tidy` 因 gin-contrib/cors v1.7.7 依赖要求 go >= 1.25 而自动升级工具链 | 固定版本 `gin@v1.9.1 cors@v1.5.0`，`GOTOOLCHAIN=local` 强制本地工具链 | `go build -o config-center .` 成功 | `feat: init backend with pinned deps` |
| BUG-002 | ✅ Closed | `frontend/TopBar.vue` | 2026-06-12 | `npm run build` 报 `Invalid end tag` | `<button>` 标签后用 `</span>` 闭合 | 改为 `</button>` | `npm run build` 成功 | `feat: init frontend` |
| BUG-003 | ✅ Closed | `frontend/ConfigTable.vue` | 2026-06-12 | `npm run build` 报 `ILLEGAL_REASSIGNMENT` | `preview` 声明为 `const` 后 `preview += '...'` | 改为 `let preview` | `npm run build` 成功 | `feat: init frontend` |
| BUG-004 | ✅ Closed | `infra/apt` | 2026-06-12 | `apt-get install golang-go` 失败 | 容器文件系统部分只读 | 手动下载 Go `go1.22.5.linux-amd64.tar.gz` 到 `/root/go` | `go version` → `go1.22.5` | `feat: init backend` |

---

## 变更日志

| 日期 | 操作 |
|------|------|
| 2026-06-12 | 创建 BUGLOG.md，迁移 AI_PROMPTS.md 中 4 个已修复 bug |
