<template>
  <div class="app-shell">
    <!-- Sidebar -->
    <Sidebar
      v-if="isDesktop"
      :configs="configs"
      :selectedEnv="filterState.env"
      :selectedStatus="filterState.status"
      @select-env="onSidebarEnv"
      @select-status="onSidebarStatus"
    />

    <!-- Main area -->
    <div class="main-area">
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

    <!-- Detail panel / drawer -->
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
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getAllConfigs } from './api/configService.js'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import ConfigTable from './components/ConfigTable.vue'
import ConfigDrawer from './components/ConfigDrawer.vue'
import { hasDraft } from './utils/configDiff.js'

// === Responsive ===
const isDesktop = ref(false)
let mq = null

onMounted(() => {
  mq = window.matchMedia('(min-width: 1024px)')
  isDesktop.value = mq.matches
  mq.addEventListener('change', (e) => { isDesktop.value = e.matches })
})

onUnmounted(() => {
  if (mq) mq.removeEventListener('change', () => {})
})

// === Data ===
const configs = ref([])
const loading = ref(true)
const refreshing = ref(false)
const lastRefresh = ref('--:--:--')
let intervalId = null

// === Filters ===
const filterState = ref({ search: '', env: '', status: '' })

// === Detail ===
const drawerVisible = ref(false)
const selectedService = ref('')
const selectedEnv = ref('')
const selectedRow = ref(null)

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

// === Methods ===
async function fetchData() {
  try {
    const data = await getAllConfigs()
    configs.value = data
    updateLastRefresh()
  } catch (e) {
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
  drawerVisible.value = true
}

function closeDetail() {
  drawerVisible.value = false
  selectedRow.value = null
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
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-app);
}

.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}
</style>
