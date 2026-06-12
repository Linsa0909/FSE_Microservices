<template>
  <aside class="sidebar">
    <div class="sb-brand">
      <svg class="sb-logo" width="22" height="22" viewBox="0 0 22 22" fill="none">
        <rect width="22" height="22" rx="5" fill="#5e6ad2"/>
        <path d="M6 11l3 3 7-7" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <span class="sb-title">配置中心</span>
    </div>

    <div class="sb-nav">
      <!-- All -->
      <button
        class="sb-item"
        :class="{ active: !selectedEnv && !selectedStatus }"
        @click="$emit('select-env', '') || $emit('select-status', '')"
      >
        <span class="sb-label">全部配置</span>
        <span class="sb-count">{{ totalCount }}</span>
      </button>

      <div class="sb-divider"></div>

      <!-- Environments -->
      <div class="sb-section-title">环境</div>
      <button
        v-for="env in envs"
        :key="env.key"
        class="sb-item"
        :class="{ active: selectedEnv === env.key }"
        @click="$emit('select-env', env.key); $emit('select-status', '')"
      >
        <span class="sb-label">{{ env.key }}</span>
        <span class="sb-count">{{ env.count }}</span>
      </button>

      <div class="sb-divider"></div>

      <!-- Status -->
      <div class="sb-section-title">状态</div>
      <button
        v-for="st in statuses"
        :key="st.key"
        class="sb-item"
        :class="{ active: selectedStatus === st.key }"
        @click="$emit('select-status', st.key); $emit('select-env', '')"
      >
        <span class="sb-dot" :class="st.key"></span>
        <span class="sb-label">{{ st.label }}</span>
        <span class="sb-count">{{ st.count }}</span>
      </button>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { hasDraft } from '../utils/configDiff.js'

const props = defineProps({
  configs: { type: Array, default: () => [] },
  selectedEnv: { type: String, default: '' },
  selectedStatus: { type: String, default: '' },
})

defineEmits(['select-env', 'select-status'])

const totalCount = computed(() => props.configs.length)

const envs = computed(() => {
  const counts = {}
  for (const c of props.configs) {
    counts[c.env] = (counts[c.env] || 0) + 1
  }
  return ['dev', 'test', 'prod'].map(k => ({ key: k, count: counts[k] || 0 }))
})

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
  min-width: 220px;
  height: 100%;
  background: var(--bg-surface);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  user-select: none;
  overflow-y: auto;
}

.sb-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
}

.sb-logo {
  flex-shrink: 0;
}

.sb-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.sb-nav {
  padding: 0 8px 16px;
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
</style>
