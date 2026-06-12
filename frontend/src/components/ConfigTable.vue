<template>
  <div class="table-wrapper">
    <div class="table-inner">
      <!-- Header -->
      <div class="table-header">
        <div class="th th-service">Service</div>
        <div class="th th-env">Env</div>
        <div class="th th-version">Ver</div>
        <div class="th th-preview">Config Preview</div>
        <div class="th th-status">Status</div>
      </div>

      <!-- Loading skeleton -->
      <div v-if="loading" class="table-body">
        <div v-for="n in 3" :key="n" class="tr shimmer">
          <div class="td th-service"><span class="shimmer-bar" style="width:60%"></span></div>
          <div class="td th-env"><span class="shimmer-bar" style="width:40%"></span></div>
          <div class="td th-version"><span class="shimmer-bar" style="width:30%"></span></div>
          <div class="td th-preview"><span class="shimmer-bar" style="width:80%"></span></div>
          <div class="td th-status"><span class="shimmer-bar" style="width:50%"></span></div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-else-if="rows.length === 0" class="empty-state">
        <div class="empty-dots">
          <svg width="80" height="80" viewBox="0 0 80 80" fill="none" opacity="0.15">
            <circle cx="20" cy="20" r="2" fill="#9CA3AF"/><circle cx="40" cy="20" r="2" fill="#9CA3AF"/>
            <circle cx="60" cy="20" r="2" fill="#9CA3AF"/><circle cx="20" cy="40" r="2" fill="#9CA3AF"/>
            <circle cx="40" cy="40" r="2" fill="#9CA3AF"/><circle cx="60" cy="40" r="2" fill="#9CA3AF"/>
            <circle cx="20" cy="60" r="2" fill="#9CA3AF"/><circle cx="40" cy="60" r="2" fill="#9CA3AF"/>
            <circle cx="60" cy="60" r="2" fill="#9CA3AF"/>
          </svg>
        </div>
        <p class="empty-title">暂无配置数据</p>
        <p class="empty-desc">后端尚未返回配置组，请确认配置中心已启动</p>
      </div>

      <!-- Rows -->
      <div v-else class="table-body">
        <div
          v-for="row in rows"
          :key="rowKey(row)"
          class="tr"
          :class="{ selected: selectedRow && selectedRow.service === row.service && selectedRow.env === row.env }"
          @click="$emit('select', row)"
        >
          <div class="td th-service">
            <span class="service-name">{{ row.service }}</span>
          </div>
          <div class="td th-env">
            <span class="env-tag">{{ row.env }}</span>
          </div>
          <div class="td th-version">
            <span class="version-num">v{{ row.publishedVersion }}</span>
          </div>
          <div class="td th-preview">
            <span class="preview-text">{{ previewText(row) }}</span>
          </div>
          <div class="td th-status">
            <StatusBadge :hasDraft="rowHasDraft(row)" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  rows: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  selectedRow: { type: Object, default: null },
})

defineEmits(['select'])

function rowKey(row) {
  return `${row.service}:${row.env}`
}

function rowHasDraft(row) {
  const d = row.draftData || {}
  const p = row.publishedData || {}
  return JSON.stringify(d) !== JSON.stringify(p)
}

function previewText(row) {
  const data = row.publishedData || {}
  const entries = Object.entries(data).slice(0, 3)
  const parts = entries.map(([k, v]) => `${k}=${v}`)
  let preview = parts.join(', ')
  if (Object.keys(data).length > 3) preview += '...'
  return preview || '—'
}
</script>

<style scoped>
.table-wrapper {
  margin: 0 24px 24px;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-input);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

.table-inner {
  width: 100%;
}

.table-header {
  display: flex;
  align-items: center;
  height: 36px;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-page);
}

.table-body {
  /* empty */
}

.th {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-weight: 500;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.th-service { width: 22%; }
.th-env { width: 10%; }
.th-version { width: 8%; }
.th-preview { flex: 1; }
.th-status { width: 160px; text-align: right; padding-right: 16px; }

.tr {
  display: flex;
  align-items: center;
  height: var(--row-height);
  padding: 0 16px;
  border-bottom: 1px solid var(--border-default);
  cursor: pointer;
  transition: background 0.1s;
}

.tr:last-child {
  border-bottom: none;
}

.tr:hover {
  background: var(--bg-hover);
}

.tr.selected {
  background: var(--bg-blue);
}

.td {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  overflow: hidden;
  white-space: nowrap;
}

.service-name {
  font-weight: 500;
}

.env-tag {
  display: inline-block;
  padding: 1px 8px;
  border-radius: var(--border-radius-tag);
  background: var(--bg-page);
  border: 1px solid var(--border-default);
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.version-num {
  font-variant-numeric: tabular-nums;
  color: var(--color-text-secondary);
}

.preview-text {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

/* Shimmer loading */
.shimmer .shimmer-bar {
  display: block;
  height: 12px;
  border-radius: 3px;
  background: linear-gradient(90deg, #eee 25%, #e0e0e0 50%, #eee 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.empty-dots {
  margin-bottom: 16px;
}

.empty-title {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: 4px;
}

.empty-desc {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}
</style>
