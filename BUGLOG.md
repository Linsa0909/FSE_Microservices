# 🐛 Bug Tracking Log

> 分布式配置中心 — 开发阶段 Bug 追踪
> 格式: | ID | 状态 | 模块 | 发现时间 | 现象 | 根因 | 修复方案 | 验证方式 | 关联 Commit |

---

## 统计

| 状态 | 数量 |
|------|------|
| ✅ Closed | 11 (bug) + 13 (lint) |
| ✅ Done | 3 (opt) |
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
| BUG-005 | ✅ Closed | `demo-service/main.go` | 2026-06-12 | 后端未启动时 `go run .` 直接 `log.Fatal` 退出 | pullConfig 无重试机制，一次失败就调用 `log.Fatalf` | pullConfig 加重试循环 (5次, 间隔1s)，main() 中 `log.Fatalf` 改为 `log.Printf` 降级启动 | 后端不启动时 demo 重试 5 次后自动以默认端口 3000 启动 | `fix: demo-service pullConfig 加重试` |
| BUG-006 | ✅ Closed | `start.sh` | 2026-06-12 | `bash start.sh` 报 `go: command not found`，后端和微服务均未启动 | WSL 中 Go 装在了 `/root/go/bin`，不在默认 PATH 中 | start.sh 开头增加 Go PATH 自动探测（`/root/go/bin`、`/usr/local/go/bin` 等 4 个常见路径 + `command -v` 校验 + 报错提示） | `bash start.sh` 找到 Go 并正常启动后端+微服务 | `fix: start.sh Go PATH auto-detect` |
| BUG-007 | ✅ Closed | `frontend/ConfigDrawer.vue` | 2026-06-12 | 点击表格行 → 右侧抽屉/面板完全空白，无任何按钮，无法操作 | `<script>` (非setup) 块定义的 `DetailBody` 子组件模板中使用了 `<ChangeLog>`，但 `ChangeLog` 只在 `<script setup>` 中 import，子组件不可见，Vue 渲染失败 | 删除非 setup 的 `<script>` 块，把 `DetailBody` 内联到主 `<template>` 中（panel + drawer 各一份） | `npm run build` 成功，点击行后抽屉显示双列对比 + 编辑/新增/删除/发布按钮 + 变更记录 | `fix: BUG-007 ConfigDrawer 重写` |
| BUG-008 | ✅ Closed | `backend/handler` | 2026-06-12 | `PUT /keys/:key` 设置 `value=""` 返回 400 `{"error":"value is required"}` | `setKeyBody.Value` 有 `binding:"required"` 标签，Gin 将空字符串视为零值拒绝 | 移除 `binding:"required"`，更新 `TestHandler_SetKey_MissingBody` 从 `{}` 改为无效 JSON 确保错误分支仍被覆盖 | `go test ./... -count=1` → 新增 `TestHandler_SetKey_EmptyValue` PASS | `test: 增强 API 并发和边界测试` |
| BUG-009 | ✅ Closed | `frontend/ConfigDrawer.vue` | 2026-06-12 | 已有键值对点击编辑按钮无法进入输入态 | `diffRows` 是 computed 生成的临时对象数组，`editing/editValue` 写在临时对象上不被 Vue 响应式系统跟踪 | 编辑状态提升为组件级 `editingKey`/`editingValue`，行根据 key 判断是否处于编辑态 | `npm run build` 通过，编辑按钮点击后可靠切换为 input | 待提交 |
| BUG-010 | ✅ Closed | `frontend/ConfigTable.vue` | 2026-06-12 | 未发布配置组显示 `v` 无版本号，最近发布时间显示 `739778 天前`，diff 文案将"新增 DB"误写为"已删除 → value" | 未发布时 `publishedVersion` 为空字段、`lastPublishedAt` 为 Go 零值时间 (0001-01-01)；`importantChanges` 对新增 key 未正确处理 | `versionLabel()` 空值返回 `—`；`isGoZeroTime()` 检测 Go 零值返回 `—`；`importantChanges` 新增 key 显示"新增"替代"已删除" | `npm run build` 通过 | 待提交 |
| BUG-011 | ✅ Closed | `frontend/App.vue` | 2026-06-12 | 待发布配置在 Dashboard 中看不到数据入口 | 缺少待发布队列摘要和快捷筛选入口 | Dashboard 增加 `pending-strip` 待发布队列（最多 4 个 chip），"查看待发布"按钮一键切换过滤 | `npm run build` 通过 | 待提交 |
| BUG-012 | ✅ Closed | `frontend/EnvironmentPanel.vue` | 2026-06-12 | 环境与推送页有新建配置组按钮 | 环境页应只展示隔离模型和预留接口，不应有配置创建入口 | 移除环境页新建配置组按钮，创建入口只留在 Dashboard 和配置管理页 | `npm run build` 通过 | 待提交 |

---

## 优化记录

| ID | 状态 | 模块 | 发现 | 优化内容 | 效果 | 验证 | Commit |
|----|------|------|------|---------|------|------|--------|
| OPT-001 | ✅ Done | `frontend/TopBar.vue` | 2026-06-12 | 缺少操作帮助入口 | 顶部栏右侧新增 ? Help 按钮 + el-dialog 操作指南 (6章节) | `npm run build` OK | `fix: BUG-007 + OPT-001` |
| OPT-002 | ✅ Done | `backend/store/` | 2026-06-12 | 版本号为单整数 | PublishedVersion int→string "1.0.0"三位, 新增 BumpVersion() | 24 tests PASS | `feat: semver + 汉化` |
| OPT-003 | ✅ Done | `frontend/` | 2026-06-12 | 界面文本为英文 | StatusBadge/Sidebar/TopBar/ConfigDrawer/ConfigTable/ChangeLog 全部汉化 | `npm run build` OK | `feat: semver + 汉化` |

## Lint 记录

| ID | 状态 | 严重度 | 文件 | 发现 | 问题 | 修复 | 验证 | Commit |
|----|------|--------|------|------|------|------|------|--------|
| LINT-001 | ✅ Closed | ⚠️ | `store/store.go` | 2026-06-12 | 缺少 package comment | 添加 package 文档注释 | `revive` 0 issues | `lint: fix 12 revive + 1 gofmt` |
| LINT-002 | ✅ Closed | ⚠️ | `handler/config.go` | 2026-06-12 | 缺少 package comment | 添加 package 文档注释 | `revive` 0 issues | `lint: fix 12 revive + 1 gofmt` |
| LINT-003 | ✅ Closed | ⚠️ | `main.go` | 2026-06-12 | 缺少 package comment | 添加 package 文档注释 | `revive` 0 issues | `lint: fix 12 revive + 1 gofmt` |
| LINT-004 | ✅ Closed | ⚠️ | `handler/config.go` | 2026-06-12 | 导出类型 `SetKeyBody` 缺少注释 | 改为未导出 `setKeyBody` | `revive` 0 issues | `lint: fix 12 revive + 1 gofmt` |
| LINT-005~012 | ✅ Closed | ⚠️ | `handler/config.go` | 2026-06-12 | 8 个导出方法注释格式不符合 Go 规范 | 全部改为 `// MethodName handles ...` 格式 | `revive` 0 issues | `lint: fix 12 revive + 1 gofmt` |
| FMT-001 | ✅ Closed | ℹ️ | `handler/handler_test.go` | 2026-06-12 | gofmt 缩进不一致 | `gofmt -w .` | 0 diffs | `lint: fix 12 revive + 1 gofmt` |

---
## 测试记录

| 日期 | 测试层级 | 用例数 | 结果 | 备注 |
|------|---------|--------|------|------|
| 2026-06-12 | UT (store) | 7 | ✅ PASS | Draft隔离/Version递增/DeepCopy/HasDraft/ChangeLog/CloneMap/DeleteKey |
| 2026-06-12 | IT (handler) | 15 | ✅ PASS | 8端点 + 完整编辑发布流程 + 10并发写入 |
| 2026-06-12 | Smoke | 17 | ✅ PASS | 编译→启动→API→编辑发布→demo拉取→清理 全流程 |

> TDD 结果: **0 个新 Bug 发现**，全部 22+17 测试 PASS

---

## 变更日志

| 日期 | 操作 |
|------|------|
| 2026-06-12 | 创建 BUGLOG.md，迁移 AI_PROMPTS.md 中 4 个已修复 bug |
| 2026-06-12 | TDD 集成测试 + 冒烟测试: 22+17 PASS, 0 新 bug |
| 2026-06-12 | Lint 首次检查 (revive+go vet+gofmt): 13 issues 全部修复, 代码 CLEAN |
| 2026-06-12 | BUG-005: demo-service pullConfig 加重试 (5次/1s), 后端未就绪时降级启动 |
| 2026-06-12 | BUG-006: start.sh Go PATH auto-detect, WSL 兼容修复 |
| 2026-06-12 | BUG-007: ConfigDrawer 重写 (去掉 DetailBody 间接层) + OPT-001: 帮助弹窗 |
| 2026-06-12 | BUG-009: ConfigDrawer 编辑状态修复 (editingKey/editingValue) + BUG-010: 版本/时间展示修复 + BUG-011: 待发布队列 + BUG-012: 环境页移除创建入口 |
