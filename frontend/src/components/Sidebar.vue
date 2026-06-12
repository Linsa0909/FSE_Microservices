<template>
  <aside class="sidebar">
    <div class="sb-nav">
      <!-- 全部 -->
      <button
        class="sb-item"
        :class="{ active: !selectedService && !selectedEnv && !selectedStatus }"
        @click="clearAll"
      >
        <span class="sb-label">全部配置组</span>
        <span class="sb-count">{{ totalCount }}</span>
      </button>

      <div class="sb-divider"></div>

      <!-- 服务列表 — 从后端 configs 数据聚合 -->
      <div class="sb-section-title">服务</div>
      <button
        v-for="svc in services"
        :key="svc.name"
        class="sb-item"
        :class="{ active: selectedService === svc.name }"
        @click="$emit('select-service', svc.name)"
      >
        <span class="sb-label">{{ svc.name }}</span>
        <span class="sb-count">{{ svc.count }}</span>
      </button>

      <div class="sb-divider"></div>

      <!-- 环境 -->
      <div class="sb-section-title">环境</div>
      <button
        v-for="env in envs"
        :key="env.key"
        class="sb-item"
        :class="{ active: selectedEnv === env.key }"
        @click="$emit('select-env', env.key)"
      >
        <span class="sb-label">{{ env.key }}</span>
        <span class="sb-count">{{ env.count }}</span>
      </button>

      <div class="sb-divider"></div>

      <!-- 状态 -->
      <div class="sb-section-title">状态</div>
      <button
        v-for="st in statuses"
        :key="st.key"
        class="sb-item"
        :class="{ active: selectedStatus === st.key }"
        @click="$emit('select-status', st.key)"
      >
        <span class="sb-dot" :class="st.key"></span>
        <span class="sb-label">{{ st.label }}</span>
        <span class="sb-count">{{ st.count }}</span>
      </button>
    </div>

    <!-- 底部说明 -->
    <div class="sb-footer">
      <p>配置组 = service + env</p>
      <p>配置项 = 组内的 key-value</p>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { hasDraft } from '../utils/configDiff.js'

const props = defineProps({
  configs: { type: Array, default: () => [] },
  selectedService: { type: String, default: '' },
  selectedEnv: { type: String, default: '' },
  selectedStatus: { type: String, default: '' },
})

const emit = defineEmits(['select-service', 'select-env', 'select-status'])

function clearAll() {
  emit('select-service', '')
  emit('select-env', '')
  emit('select-status', '')
}

// ── 总数 = 配置组数 ──
const totalCount = computed(() => props.configs.length)

// ── 服务列表 — 从后端数据聚合 ──
const services = computed(() => {
  const map = {}
  for (const c of props.configs) {
    map[c.service] = (map[c.service] || 0) + 1
  }
  return Object.entries(map)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, count]) => ({ name, count }))
})

// ── 环境 — 统计每个 env 下的配置组数 ──
const envs = computed(() => {
  const counts = {}
  for (const c of props.configs) {
    counts[c.env] = (counts[c.env] || 0) + 1
  }
  return ['dev', 'test', 'prod'].map(k => ({ key: k, count: counts[k] || 0 }))
})

// ── 状态 — 已发布 vs 待发布 ──
const statuses = computed(() => {
  let pub = 0, pen = 0
  for (const c of props.configs) {
    if (hasDraft(c)) pen++
    else pub++
  }
  return [
    { key: 'published', label: '已发布', count: pub },
    { key: 'pending', label: '待发布', count: pen },
  ]
})
</script>

<style scoped>
.sidebar {
  width: 220px;
  min-width: 200px;
  height: 100%;
  background: var(--bg-surface);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  user-select: none;
  overflow-y: auto;
}

.sb-nav {
  padding: 12px 8px 16px;
  flex: 1;
}

.sb-divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 8px 8px;
}

.sb-section-title {
  padding: 4px 8px 6px;
  font-size: var(--font-size-xs);
  font-weight: 500;
  color: var(--text-placeholder);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.sb-item {
  display: flex;
  align-items: center;
  width: 100%;
  height: 32px;
  padding: 0 8px;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
  gap: 8px;
}

.sb-item:hover {
  background: var(--bg-subtle);
  color: var(--text-primary);
}

.sb-item.active {
  background: var(--bg-blue);
  color: var(--color-primary);
}

.sb-label {
  flex: 1;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sb-count {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  font-variant-numeric: tabular-nums;
  min-width: 20px;
  text-align: right;
}

.sb-item.active .sb-count {
  color: var(--color-primary);
}

.sb-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sb-dot.published { background: var(--color-published); }
.sb-dot.pending { background: var(--color-pending); }

/* ── 底部术语说明 ── */
.sb-footer {
  padding: 10px 12px;
  border-top: 1px solid var(--border-subtle);
  font-size: 11px;
  color: var(--text-placeholder);
  line-height: 1.6;
}
.sb-footer p { margin: 0; }
</style>
