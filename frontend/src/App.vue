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
        <div class="overview-actions">
          <button class="refresh-button" @click="fetchData">
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" :class="{ spinning: refreshing }">
              <path d="M13.65 2.35A7.96 7.96 0 008 0a8 8 0 100 16 7.96 7.96 0 005.65-2.35" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            刷新
          </button>
          <button class="primary-button" @click="openCreateDialog()">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            新建配置组
          </button>
        </div>
      </section>

      <!-- Dashboard -->
      <section class="dashboard">
        <div class="dashboard-head">
          <div>
            <h2>Dashboard</h2>
            <p>配置中心运行概览、发布风险和环境覆盖情况。</p>
          </div>
          <button class="secondary-button" @click="showPendingOnly">
            查看待发布
          </button>
        </div>

        <div class="stats-grid">
          <article class="stat-card">
            <span class="stat-label">配置组</span>
            <strong>{{ overviewStats.total }}</strong>
            <small>service + env 隔离</small>
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
          <article class="stat-card">
            <span class="stat-label">刷新状态</span>
            <strong>{{ lastRefresh }}</strong>
            <small>每 5 秒自动同步</small>
          </article>
        </div>

        <div class="pending-strip">
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

      <!-- Tab navigation -->
      <nav class="workspace-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-button"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <svg v-if="tab.key === 'configs'" width="14" height="14" viewBox="0 0 16 16" fill="none">
            <rect x="2" y="2" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.3"/>
            <rect x="9" y="2" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.3"/>
            <rect x="2" y="9" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.3"/>
            <rect x="9" y="9" width="5" height="5" rx="1" stroke="currentColor" stroke-width="1.3"/>
          </svg>
          <svg v-else-if="tab.key === 'services'" width="14" height="14" viewBox="0 0 16 16" fill="none">
            <circle cx="5" cy="4" r="2" stroke="currentColor" stroke-width="1.3"/>
            <circle cx="11" cy="4" r="2" stroke="currentColor" stroke-width="1.3"/>
            <circle cx="5" cy="12" r="2" stroke="currentColor" stroke-width="1.3"/>
            <circle cx="11" cy="12" r="2" stroke="currentColor" stroke-width="1.3"/>
            <line x1="5" y1="6" x2="5" y2="10" stroke="currentColor" stroke-width="1"/>
            <line x1="11" y1="6" x2="11" y2="10" stroke="currentColor" stroke-width="1"/>
            <line x1="7" y1="4" x2="9" y2="4" stroke="currentColor" stroke-width="1"/>
            <line x1="7" y1="12" x2="9" y2="12" stroke="currentColor" stroke-width="1"/>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 16 16" fill="none">
            <rect x="2" y="2" width="12" height="12" rx="2" stroke="currentColor" stroke-width="1.3"/>
            <line x1="2" y1="7" x2="14" y2="7" stroke="currentColor" stroke-width="1"/>
            <line x1="7" y1="7" x2="7" y2="14" stroke="currentColor" stroke-width="1"/>
          </svg>
          {{ tab.label }}
        </button>
      </nav>

      <!-- Tab: 配置管理 -->
      <section v-if="activeTab === 'configs'" class="workspace config-workspace">
        <Sidebar
          v-if="isDesktop"
          :configs="configs"
          :selectedEnv="filterState.env"
          :selectedStatus="filterState.status"
          @select-env="onSidebarEnv"
          @select-status="onSidebarStatus"
        />

        <div class="config-main">
          <div class="workspace-title-row">
            <div>
              <h2>配置管理</h2>
              <p>配置新增、修改、删除和发布只在这里完成，微服务状态页不承载配置编辑。</p>
            </div>
            <button class="secondary-button" @click="openCreateDialog()">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              新建配置组
            </button>
          </div>

          <TopBar
            :search="filterState.search"
            :env="filterState.env"
            :status="filterState.status"
            :lastRefresh="lastRefresh"
            :refreshing="refreshing"
            @manual-refresh="fetchData"
            @update:search="filterState.search = $event"
            @update:env="filterState.env = $event"
            @update:status="filterState.status = $event"
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
      </section>

      <!-- Tab: 微服务状态 -->
      <section v-else-if="activeTab === 'services'" class="workspace">
        <ServiceStatus :configs="configs" />
      </section>

      <!-- Tab: 环境与推送 -->
      <section v-else class="workspace">
        <EnvironmentPanel :configs="configs" />
      </section>
    </div>

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
.secondary-button,
.refresh-button {
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

.secondary-button,
.refresh-button {
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.82);
}

.secondary-button:hover,
.refresh-button:hover {
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

/* Tab navigation */
.workspace-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.74);
  padding: 4px;
  width: fit-content;
  flex-shrink: 0;
}

.tab-button {
  height: 34px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 0 14px;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 600;
  cursor: pointer;
}

.tab-button:hover {
  color: var(--text-primary);
  background: var(--bg-subtle);
}

.tab-button.active {
  color: var(--color-primary);
  background: var(--bg-surface);
  border-color: var(--border-subtle);
  box-shadow: var(--shadow-sm);
}

/* Workspace */
.workspace {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.config-workspace {
  display: flex;
}

.config-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.workspace-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 16px 14px;
  border-bottom: 1px solid var(--border-subtle);
  background: linear-gradient(180deg, #fff, #fbfcff);
}

.workspace-title-row h2 {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.workspace-title-row p {
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
  .workspace-tabs { width: 100%; overflow-x: auto; }
  .tab-button { flex: 1; white-space: nowrap; }
  .workspace-title-row { flex-direction: column; }
}
</style>
