<template>
  <div class="app-shell">
    <div class="main-area">
      <!-- Overview header -->
      <section class="overview-header">
        <div class="overview-copy">
          <div class="eyebrow">Distributed Config Center</div>
          <h1>分布式配置中心</h1>
          <p>按服务和环境隔离配置，支持草稿编辑、发布生效、示例微服务启动拉取和推送接口预留。</p>
        </div>
      </section>

      <!-- Dashboard -->
      <section class="dashboard">
        <div class="dashboard-head">
          <div>
            <h2>Dashboard</h2>
            <p>配置中心运行概览、发布风险和环境覆盖情况。</p>
          </div>
          <div class="dashboard-head-actions">
            <span class="dash-time">{{ lastRefresh }}</span>
            <button class="secondary-button-sm" @click="fetchData" title="手动刷新">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none" :class="{ spinning: refreshing }">
                <path d="M13.65 2.35A7.96 7.96 0 008 0a8 8 0 100 16 7.96 7.96 0 005.65-2.35" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
            <button class="secondary-button" @click="showPendingOnly">
              查看待发布
            </button>
            <button class="secondary-button-sm" @click="showHelp = true" title="帮助">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
                <circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.3"/>
                <path d="M6.5 6a1.5 1.5 0 012.8-.7M8 9.5V12" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/>
              </svg>
            </button>
            <button class="collapse-btn" @click="dashboardCollapsed = !dashboardCollapsed" :title="dashboardCollapsed ? '展开 Dashboard' : '收起 Dashboard'">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none" :class="{ rotated: dashboardCollapsed }">
                <path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>
        </div>

        <div v-show="!dashboardCollapsed" class="stats-grid">
          <article class="stat-card">
            <span class="stat-label">配置组</span>
            <strong>{{ overviewStats.total }}</strong>
            <small>service + env 隔离</small>
          </article>
          <article class="stat-card published-card">
            <span class="stat-label">已发布</span>
            <strong>{{ overviewStats.published }}</strong>
            <small>{{ overviewStats.published > 0 ? '所有配置项已同步' : '暂无已发布配置' }}</small>
          </article>
          <article class="stat-card warning-card">
            <span class="stat-label">待发布</span>
            <strong>{{ overviewStats.pending }}</strong>
            <small>{{ overviewStats.pending > 0 ? '需要发布后微服务才可读取' : '无未发布草稿' }}</small>
          </article>
          <article class="stat-card">
            <span class="stat-label">环境数</span>
            <strong>{{ overviewStats.envs }}</strong>
            <small>dev / test / prod / 自定义</small>
          </article>
        </div>

        <div v-show="!dashboardCollapsed" class="pending-strip">
          <span class="pending-title">待发布队列</span>
          <template v-if="pendingConfigs.length > 0">
            <button
              v-for="item in pendingConfigs.slice(0, 4)"
              :key="`${item.service}:${item.env}`"
              class="pending-chip"
              @click="openDetail(item)"
            >
              {{ item.service }} / {{ item.env }}
            </button>
          </template>
          <span v-else class="pending-empty">当前没有待发布配置组</span>
        </div>
      </section>

      <!-- Workspace box: tabs + content inside the same border -->
      <section class="workspace" :class="{ 'config-workspace': activeTab === 'configs' }">
        <nav class="workspace-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="tab-button"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
          <div class="tab-spacer"></div>
          <button class="secondary-button-sm" @click="openCreateDialog()">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            新建配置组
          </button>
        </nav>

        <!-- Tab: 配置管理 -->
        <div v-if="activeTab === 'configs'" class="workspace-body config-layout">
          <Sidebar
            v-if="isDesktop"
            :configs="configs"
            :selectedEnv="filterState.env"
            :selectedStatus="filterState.status"
            @select-env="onSidebarEnv"
            @select-status="onSidebarStatus"
          />

          <div class="config-main">
            <div class="panel-head">
              <div>
                <h2>配置管理</h2>
                <p>配置新增、修改、删除和发布只在这里完成，微服务状态页不承载配置编辑。</p>
              </div>
            </div>
            <TopBar
              :search="filterState.search"
              @update:search="filterState.search = $event"
            />
            <ConfigTable
              :rows="filteredConfigs"
              :loading="loading"
              :selectedRow="selectedRow"
              @select="openDetail"
            />
          </div>

          <template v-if="selectedRow">
            <ConfigDrawer
              v-if="isDesktop"
              :service="selectedService"
              :env="selectedEnv"
              :panel="true"
              @close="closeDetail"
              @published="fetchData"
            />
            <ConfigDrawer
              v-else
              v-model="drawerVisible"
              :service="selectedService"
              :env="selectedEnv"
              @close="closeDetail"
              @published="fetchData"
            />
          </template>
        </div>

        <!-- Tab: 微服务状态 -->
        <div v-else-if="activeTab === 'services'" class="workspace-body">
          <ServiceStatus :configs="configs" />
        </div>

        <!-- Tab: 环境与推送 -->
        <div v-else class="workspace-body">
          <EnvironmentPanel :configs="configs" />
        </div>
      </section>
    </div>

    <!-- Help dialog -->
    <el-dialog v-model="showHelp" title="操作指南" width="480px">
      <div class="help-content">
        <h4>Dashboard</h4>
        <p>页面顶部 Dashboard 展示配置组、待发布、已发布、环境数和刷新状态；点击待发布队列可直接打开对应配置组。</p>
        <h4>配置管理</h4>
        <p>进入配置管理页 → 选择环境/状态或搜索服务 → 点击配置组行 → 右侧展开 Key-Value 编辑面板。</p>
        <h4>新建配置组</h4>
        <p>点击新建配置组 → 填写服务名和环境 → 从配置模板选择常用 Key/Value；需要特殊配置时选择自定义后手动输入。</p>
        <h4>新增配置项</h4>
        <p>展开详情面板 → 在 <strong>新增配置项</strong> 区域填写 Key 和 Value → 点击添加或按回车。</p>
        <h4>修改 / 删除</h4>
        <p>在详情面板的 <strong>草稿</strong> 列中点击值编辑，或点击右侧 <strong>删除图标</strong> 移除。</p>
        <h4>发布配置</h4>
        <p>编辑完成后点击底部 <strong>发布</strong> 按钮 → 草稿同步到已发布配置 → 版本号 +1。</p>
        <p class="help-note">未做任何修改时发布按钮禁用；有未发布变更时高亮显示。</p>
        <h4>自动刷新</h4>
        <p>页面每 <strong>5 秒</strong> 自动拉取后端最新数据。</p>
      </div>
      <template #footer>
        <el-button @click="showHelp = false">知道了</el-button>
      </template>
    </el-dialog>

    <!-- Create config group dialog -->
    <el-dialog v-model="createVisible" title="新建配置组" width="480px" class="create-dialog">
      <div class="create-form">
        <label>
          <span>服务名</span>
          <el-input v-model="newGroup.service" placeholder="order-service" />
        </label>
        <label>
          <span>环境</span>
          <el-input v-model="newGroup.env" placeholder="dev / test / prod" />
        </label>
        <label>
          <span>配置模板</span>
          <el-select
            v-model="newGroup.template"
            placeholder="选择常用配置项"
            @change="applyTemplate"
          >
            <el-option
              v-for="tpl in configTemplates"
              :key="tpl.key"
              :label="tpl.label"
              :value="tpl.key"
            />
          </el-select>
        </label>
        <label>
          <span>初始 Key</span>
          <el-input v-model="newGroup.key" placeholder="db.url" />
        </label>
        <label>
          <span>初始 Value</span>
          <el-input v-model="newGroup.value" placeholder="localhost:3306" />
        </label>
        <p class="create-hint">后端通过 PUT key 自动创建 service + env 配置组；发布前只写入 DraftData。</p>
      </div>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createConfigGroup">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAllConfigs, setKey } from './api/configService.js'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import ConfigTable from './components/ConfigTable.vue'
import ConfigDrawer from './components/ConfigDrawer.vue'
import ServiceStatus from './components/ServiceStatus.vue'
import EnvironmentPanel from './components/EnvironmentPanel.vue'
import { hasDraft } from './utils/configDiff.js'

// === Responsive ===
const isDesktop = ref(false)
let mq = null
let mqHandler = null

onMounted(() => {
  mq = window.matchMedia('(min-width: 1100px)')
  isDesktop.value = mq.matches
  mqHandler = (e) => { isDesktop.value = e.matches }
  mq.addEventListener('change', mqHandler)
})

onUnmounted(() => {
  if (mq && mqHandler) mq.removeEventListener('change', mqHandler)
})

// === Data ===
const configs = ref([])
const loading = ref(true)
const refreshing = ref(false)
const lastRefresh = ref('--:--:--')
let intervalId = null

// === Filters ===
const filterState = ref({ search: '', env: '', status: '' })

// === Tabs ===
const activeTab = ref('configs')
const tabs = [
  { key: 'configs', label: '配置管理' },
  { key: 'services', label: '微服务状态' },
  { key: 'environment', label: '环境与推送' },
]

// === Detail ===
const drawerVisible = ref(false)
const selectedService = ref('')
const selectedEnv = ref('')
const selectedRow = ref(null)

// === Create dialog ===
const createVisible = ref(false)
const dashboardCollapsed = ref(false)
const showHelp = ref(false)
const creating = ref(false)
const newGroup = ref({ service: '', env: 'dev', template: 'db.url', key: 'db.url', value: 'localhost:3306' })
const configTemplates = [
  { key: 'db.url', label: '数据库连接 db.url', value: 'localhost:3306' },
  { key: 'server.port', label: '服务端口 server.port', value: '3000' },
  { key: 'log.level', label: '日志级别 log.level', value: 'debug' },
  { key: 'redis.url', label: 'Redis 地址 redis.url', value: 'localhost:6379' },
  { key: 'custom', label: '自定义 Key / Value', value: '' },
]

// === Computed ===
const filteredConfigs = computed(() => {
  let list = configs.value
  const f = filterState.value

  if (f.search) {
    const q = f.search.toLowerCase()
    list = list.filter(c => c.service.toLowerCase().includes(q))
  }
  if (f.env) {
    list = list.filter(c => c.env === f.env)
  }
  if (f.status) {
    list = list.filter(c => {
      const draft = hasDraft(c)
      return f.status === 'pending' ? draft : !draft
    })
  }

  return list
})

const overviewStats = computed(() => {
  const envSet = new Set(configs.value.map(c => c.env).filter(Boolean))
  const pending = configs.value.filter(c => hasDraft(c)).length
  return {
    total: configs.value.length,
    pending,
    published: configs.value.length - pending,
    envs: envSet.size,
  }
})

const pendingConfigs = computed(() => configs.value.filter(c => hasDraft(c)))

// === Methods ===
async function fetchData() {
  refreshing.value = true
  try {
    const data = await getAllConfigs()
    configs.value = data
    updateLastRefresh()
  } catch {
    // Silent — shown as empty state
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function updateLastRefresh() {
  const now = new Date()
  lastRefresh.value = now.toLocaleTimeString('zh-CN', { hour12: false })
}

function onSidebarEnv(env) {
  filterState.value.env = env
}

function onSidebarStatus(status) {
  filterState.value.status = status
}

function openDetail(row) {
  selectedService.value = row.service
  selectedEnv.value = row.env
  selectedRow.value = row
  activeTab.value = 'configs'
  drawerVisible.value = true
}

function closeDetail() {
  drawerVisible.value = false
  selectedRow.value = null
}

function openCreateDialog(seed = {}) {
  const template = seed.template || 'db.url'
  const tpl = configTemplates.find(item => item.key === template) || configTemplates[0]
  newGroup.value = {
    service: seed.service || '',
    env: seed.env || 'dev',
    template,
    key: seed.key || tpl.key,
    value: seed.value ?? tpl.value,
  }
  createVisible.value = true
}

function applyTemplate(value) {
  const tpl = configTemplates.find(item => item.key === value)
  if (!tpl || tpl.key === 'custom') return
  newGroup.value.key = tpl.key
  newGroup.value.value = tpl.value
}

function showPendingOnly() {
  activeTab.value = 'configs'
  filterState.value.status = 'pending'
  filterState.value.env = ''
  closeDetail()
}

async function createConfigGroup() {
  const service = newGroup.value.service.trim()
  const env = newGroup.value.env.trim()
  const key = newGroup.value.key.trim()
  const value = newGroup.value.value

  if (!service || !env || !key) {
    ElMessage.warning('请填写服务名、环境和初始 Key')
    return
  }

  creating.value = true
  try {
    await setKey(service, env, key, value)
    await fetchData()
    createVisible.value = false
    activeTab.value = 'configs'
    openDetail({ service, env })
    ElMessage.success('配置组已创建，发布后微服务可读取 PublishedData')
  } catch {
    ElMessage.error('创建配置组失败')
  } finally {
    creating.value = false
  }
}

// === Lifecycle ===
onMounted(() => {
  fetchData()
  intervalId = setInterval(fetchData, 5000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})
</script>

<style scoped>
.app-shell {
  height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at top left, rgba(31, 111, 235, 0.08), transparent 320px),
    linear-gradient(180deg, #f7faff 0%, var(--bg-app) 42%);
}

.main-area {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  padding: 22px 24px 24px;
  gap: 14px;
}

/* Overview header */
.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  flex-shrink: 0;
}

.eyebrow {
  color: var(--color-primary);
  font-size: var(--font-size-xs);
  font-weight: 700;
  letter-spacing: 0;
  margin-bottom: 6px;
}

.overview-copy h1 {
  color: var(--text-primary);
  font-size: 26px;
  line-height: 1.2;
  font-weight: 720;
  margin-bottom: 6px;
}

.overview-copy p {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.7;
  max-width: 720px;
}

.overview-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.primary-button,
.secondary-button {
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 0 13px;
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.14s;
}

.primary-button {
  border: 1px solid var(--color-primary);
  color: #fff;
  background: var(--color-primary);
  box-shadow: 0 8px 18px rgba(31, 111, 235, 0.18);
}

.primary-button:hover {
  background: var(--color-primary-hover);
  border-color: var(--color-primary-hover);
}

.secondary-button {
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.82);
}

.secondary-button:hover {
  color: var(--text-primary);
  border-color: var(--border-hover);
  background: var(--bg-surface);
}

/* Dashboard */
.dashboard {
  padding: 14px;
  border: 1px solid rgba(214, 220, 230, 0.86);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.78);
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
}

.dashboard-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}

.dashboard-head h2 {
  font-size: 17px;
  font-weight: 720;
  color: var(--text-primary);
  margin-bottom: 3px;
}

.dashboard-head p {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.dashboard-head-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.dash-time {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  font-variant-numeric: tabular-nums;
  margin-right: 2px;
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.12s;
  margin-left: 4px;
}

.collapse-btn:hover {
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.collapse-btn svg {
  transition: transform 0.2s;
}

.collapse-btn svg.rotated {
  transform: rotate(-90deg);
}

.published-card {
  border-color: rgba(33, 154, 128, 0.28);
  background: linear-gradient(180deg, #fff, var(--color-published-bg));
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.stat-card {
  min-height: 92px;
  padding: 14px 16px;
  border: 1px solid rgba(214, 220, 230, 0.86);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.88);
  box-shadow: var(--shadow-sm);
}

.warning-card {
  border-color: rgba(176, 137, 60, 0.28);
  background: linear-gradient(180deg, #fff, var(--color-pending-bg));
}

.stat-label {
  display: block;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  margin-bottom: 8px;
}

.stat-card strong {
  display: block;
  color: var(--text-primary);
  font-size: 24px;
  line-height: 1.1;
  font-weight: 720;
  font-variant-numeric: tabular-nums;
}

.stat-card small {
  display: block;
  margin-top: 8px;
  color: var(--text-placeholder);
  font-size: var(--font-size-xs);
}

.pending-strip {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
  min-height: 34px;
}

.pending-title {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: 650;
  margin-right: 2px;
}

.pending-chip {
  height: 26px;
  padding: 0 9px;
  border: 1px solid rgba(176, 137, 60, 0.25);
  border-radius: 999px;
  background: var(--color-pending-bg);
  color: var(--color-pending);
  font-family: var(--font-family);
  font-size: var(--font-size-xs);
  font-weight: 600;
  cursor: pointer;
}

.pending-chip:hover {
  border-color: var(--color-pending);
}

.pending-empty {
  color: var(--text-placeholder);
  font-size: var(--font-size-xs);
}

/* Tab navigation — inside workspace as header */
.workspace-tabs {
  display: flex;
  align-items: center;
  gap: 0;
  border-bottom: 1px solid var(--border-subtle);
  background: #fbfbfc;
  padding: 0 8px;
  flex-shrink: 0;
}

.tab-button {
  height: 38px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 14px;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 0;
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: color 0.12s, border-color 0.12s;
}

.tab-button:hover {
  color: var(--text-primary);
}

.tab-button.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
  font-weight: 600;
}

.tab-spacer { flex: 1; }

.secondary-button-sm {
  height: 28px;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 0 10px;
  margin-right: 4px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-size-xs);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.12s;
  white-space: nowrap;
}

.secondary-button-sm:hover {
  border-color: var(--border-hover);
  color: var(--text-primary);
}

/* Workspace — contains tabs + body */
.workspace {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.workspace-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.config-layout {
  display: flex;
}

.config-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px 12px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.panel-head h2 {
  font-size: 16px;
  font-weight: 650;
  color: var(--text-primary);
  margin-bottom: 3px;
}

.panel-head p {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

/* Create dialog */
.create-form {
  display: grid;
  gap: 13px;
}

.create-form label {
  display: grid;
  gap: 6px;
}

.create-form label span {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--text-primary);
}

.create-hint {
  margin-top: 2px;
  color: var(--text-placeholder);
  font-size: var(--font-size-xs);
  line-height: 1.6;
}

/* Help dialog */
.help-content { font-family: var(--font-family); color: var(--text-primary); }
.help-content h4 { font-size: var(--font-size-base); font-weight: 600; margin: 16px 0 4px; color: var(--text-primary); }
.help-content h4:first-child { margin-top: 0; }
.help-content p { font-size: var(--font-size-sm); color: var(--text-secondary); margin: 2px 0 8px; line-height: 1.6; }
.help-note { color: var(--text-placeholder) !important; font-style: italic; }

/* Animations */
.spinning {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Responsive */
@media (max-width: 1120px) {
  .stats-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 760px) {
  .main-area { padding: 16px; }
  .overview-header { flex-direction: column; }
  .overview-actions { width: 100%; }
  .overview-actions > button { flex: 1; }
  .stats-grid { grid-template-columns: 1fr; }
  .dashboard-head { flex-direction: column; }
  .pending-strip { flex-wrap: wrap; }
  .tab-button { white-space: nowrap; }
}
</style>
