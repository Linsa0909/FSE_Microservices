<template>
  <div class="app">
    <TopBar
      :lastRefresh="lastRefresh"
      :refreshing="refreshing"
      @manual-refresh="fetchData"
    />
    <div class="page-content">
      <ConfigFilters @filter="onFilter" />
      <ConfigTable
        :rows="filteredConfigs"
        :loading="loading"
        :selectedRow="selectedRow"
        @select="openDrawer"
      />
    </div>
    <ConfigDrawer
      v-model="drawerVisible"
      :service="selectedService"
      :env="selectedEnv"
      @close="drawerVisible = false"
      @published="fetchData"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getAllConfigs } from './api/configService.js'
import TopBar from './components/TopBar.vue'
import ConfigFilters from './components/ConfigFilters.vue'
import ConfigTable from './components/ConfigTable.vue'
import ConfigDrawer from './components/ConfigDrawer.vue'

// Data
const configs = ref([])
const loading = ref(true)
const refreshing = ref(false)
const lastRefresh = ref('--:--:--')
let intervalId = null

// Filters
const filterState = ref({ search: '', env: '', status: '' })

// Drawer
const drawerVisible = ref(false)
const selectedService = ref('')
const selectedEnv = ref('')
const selectedRow = ref(null)

// Computed
const filteredConfigs = computed(() => {
  let list = configs.value
  const f = filterState.value

  if (f.search) {
    list = list.filter(c => c.service.toLowerCase().includes(f.search))
  }
  if (f.env) {
    list = list.filter(c => c.env === f.env)
  }
  if (f.status) {
    list = list.filter(c => {
      const hasDraft = JSON.stringify(c.draftData || {}) !== JSON.stringify(c.publishedData || {})
      return f.status === 'pending' ? hasDraft : !hasDraft
    })
  }

  return list
})

// Methods
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

function onFilter(f) {
  filterState.value = f
}

function openDrawer(row) {
  selectedService.value = row.service
  selectedEnv.value = row.env
  selectedRow.value = row
  drawerVisible.value = true
}

// Lifecycle
onMounted(() => {
  fetchData()
  intervalId = setInterval(fetchData, 5000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})
</script>

<style scoped>
.app {
  min-height: 100vh;
  background: var(--bg-page);
}

.page-content {
  max-width: 1200px;
  margin: 0 auto;
  padding-top: 8px;
}
</style>
