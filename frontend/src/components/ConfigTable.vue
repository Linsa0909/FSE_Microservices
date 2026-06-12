<template>
  <div class="cl-table-wrapper">
    <div class="cl-table">
      <!-- Header -->
      <div class="cl-th">
        <div class="cl-th-cell svc">Service / Env</div>
        <div class="cl-th-cell ver">Ver</div>
        <div class="cl-th-cell preview">Config Preview</div>
        <div class="cl-th-cell changes">变更</div>
        <div class="cl-th-cell time">Last Published</div>
        <div class="cl-th-cell status">Status</div>
      </div>

      <!-- Loading shimmer -->
      <div v-if="loading" class="cl-body">
        <div v-for="n in 3" :key="n" class="cl-row shimmer">
          <div class="cl-cell svc"><span class="shim" style="width:60%"></span></div>
          <div class="cl-cell ver"><span class="shim" style="width:30%"></span></div>
          <div class="cl-cell preview"><span class="shim" style="width:80%"></span></div>
          <div class="cl-cell changes"><span class="shim" style="width:40%"></span></div>
          <div class="cl-cell time"><span class="shim" style="width:50%"></span></div>
          <div class="cl-cell status"><span class="shim" style="width:60%"></span></div>
        </div>
      </div>

      <!-- Empty -->
      <div v-else-if="rows.length === 0" class="cl-empty-state">
        <div class="empty-dots">
          <svg width="60" height="60" viewBox="0 0 80 80" fill="none" opacity="0.12">
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
      <div v-else class="cl-body">
        <div
          v-for="row in rows"
          :key="rowKey(row)"
          class="cl-row"
          :class="{ selected: isSelected(row) }"
          @click="$emit('select', row)"
        >
          <div class="cl-cell svc">
            <span class="svc-name">{{ row.service }}</span>
            <span class="svc-env">{{ row.env }}</span>
          </div>
          <div class="cl-cell ver">
            <span class="ver-num">v{{ row.publishedVersion }}</span>
          </div>
          <div class="cl-cell preview">
            <span class="preview-text">{{ previewText(row) }}</span>
          </div>
          <div class="cl-cell changes">
            <span v-if="rowHasDraft(row)" class="changes-badge">{{ countDiffStr(row) }}</span>
            <span v-else class="no-changes">—</span>
          </div>
          <div class="cl-cell time">
            <span class="time-text">{{ formatRelative(row.lastPublishedAt) }}</span>
          </div>
          <div class="cl-cell status">
            <StatusBadge :hasDraft="rowHasDraft(row)" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import StatusBadge from './StatusBadge.vue'
import { hasDraft, countDiff } from '../utils/configDiff.js'

const props = defineProps({
  rows: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  selectedRow: { type: Object, default: null },
})

defineEmits(['select'])

function rowKey(row) {
  return `${row.service}:${row.env}`
}

function isSelected(row) {
  return props.selectedRow &&
    props.selectedRow.service === row.service &&
    props.selectedRow.env === row.env
}

function rowHasDraft(row) {
  return hasDraft(row)
}

function countDiffStr(row) {
  const c = countDiff(row)
  const parts = []
  if (c.added) parts.push(`+${c.added}`)
  if (c.modified) parts.push(`~${c.modified}`)
  if (c.deleted) parts.push(`-${c.deleted}`)
  return `${c.total} 项变更`
}

function previewText(row) {
  const pub = row.publishedData || {}
  const draft = row.draftData || {}
  const diff = countDiff(row)

  if (diff.total > 0) {
    const changes = importantChanges(pub, draft).slice(0, 2)
    const suffix = changes.length ? ` · ${changes.join(' · ')}` : ''
    return `${diff.total} 项待发布${suffix}`
  }

  const summary = importantEntries(pub).slice(0, 3)
  if (summary.length > 0) return summary.join(' · ')

  const total = Object.keys(pub).length
  if (total === 0) return '暂无配置项'
  return `${total} 个配置项`
}

const KEY_LABELS = [
  { keys: ['db.url', 'datasource.url', 'database.url'], label: 'DB' },
  { keys: ['server.port', 'port'], label: 'Port' },
  { keys: ['log.level', 'logging.level'], label: 'Log' },
  { keys: ['redis.host', 'redis.url'], label: 'Redis' },
  { keys: ['nacos.addr', 'config.addr'], label: 'Config' },
]

function importantEntries(data) {
  return KEY_LABELS.flatMap(({ keys, label }) => {
    const key = keys.find(k => hasOwn(data, k))
    return key ? [`${label} ${data[key]}`] : []
  })
}

function importantChanges(pub, draft) {
  return KEY_LABELS.flatMap(({ keys, label }) => {
    const key = keys.find(k => hasOwn(pub, k) || hasOwn(draft, k))
    if (!key || pub[key] === draft[key]) return []
    return [`${label} ${valueText(pub[key])} → ${valueText(draft[key])}`]
  })
}

function hasOwn(obj, key) {
  return Object.prototype.hasOwnProperty.call(obj, key)
}

function valueText(value) {
  return value === undefined ? '已删除' : value || '空值'
}

function formatRelative(t) {
  if (!t) return '—'
  const diff = Date.now() - new Date(t).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  return `${days} 天前`
}
</script>

<style scoped>
.cl-table-wrapper {
  flex: 1;
  overflow-y: auto;
  background: var(--bg-surface);
}

.cl-table {
  min-width: 100%;
}

/* Header */
.cl-th {
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-subtle);
  position: sticky;
  top: 0;
  z-index: 1;
}

.cl-th-cell {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  font-weight: 500;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.svc { width: 24%; min-width: 140px; }
.ver { width: 7%; min-width: 40px; text-align: center; }
.preview { flex: 1; min-width: 120px; }
.changes { width: 12%; min-width: 80px; }
.time { width: 12%; min-width: 70px; }
.status { width: 14%; min-width: 90px; text-align: right; padding-right: 8px; }

/* Rows */
.cl-row {
  display: flex;
  align-items: center;
  height: var(--row-height);
  padding: 0 12px;
  border-bottom: 1px solid var(--border-subtle);
  border-left: 3px solid transparent;
  cursor: pointer;
  transition: background 0.1s;
}

.cl-row:hover {
  background: var(--bg-hover);
}

.cl-row.selected {
  background: var(--bg-blue);
  border-left-color: var(--color-primary);
}

.cl-row:last-child {
  border-bottom: none;
}

/* Cells */
.cl-cell {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  overflow: hidden;
  white-space: nowrap;
}

.svc-name {
  font-weight: 500;
  color: var(--text-primary);
}

.svc-env {
  display: inline-block;
  margin-left: 6px;
  padding: 0 6px;
  border-radius: var(--radius-sm);
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.ver-num {
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
}

.preview-text {
  color: var(--text-secondary);
}

.changes-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: var(--radius-sm);
  background: var(--color-pending-bg);
  color: var(--color-pending);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

.no-changes {
  color: var(--text-placeholder);
}

.time-text {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
}

/* Shimmer */
.shim {
  display: block;
  height: 10px;
  border-radius: 3px;
  background: linear-gradient(90deg, #eee 25%, #e0e0e0 50%, #eee 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* Empty */
.cl-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.empty-dots { margin-bottom: 12px; }

.empty-title {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.empty-desc {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}
</style>
