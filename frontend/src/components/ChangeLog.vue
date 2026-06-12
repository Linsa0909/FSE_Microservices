<template>
  <div class="cl-wrapper">
    <div class="cl-header">
      <h4>Activity</h4>
      <span class="cl-count">{{ logs.length }}</span>
    </div>
    <div class="cl-list" v-if="logs.length > 0">
      <div v-for="(log, i) in logs" :key="i" class="cl-item">
        <span class="cl-dot" :class="'dot-' + actionClass(log.action)"></span>
        <span class="cl-operator">{{ log.operator }}</span>
        <span class="cl-action" :class="'act-' + actionClass(log.action)">{{ log.action }}</span>
        <code v-if="log.key" class="cl-key">{{ log.key }}</code>
        <span class="cl-spacer"></span>
        <span class="cl-time">{{ formatTime(log.time) }}</span>
      </div>
    </div>
    <div v-else class="cl-empty">No activity yet</div>
  </div>
</template>

<script setup>
defineProps({
  logs: { type: Array, default: () => [] },
})

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

function actionClass(action) {
  const m = { '新增': 'add', '修改': 'modify', '删除': 'delete', '发布': 'publish' }
  return m[action] || 'default'
}
</script>

<style scoped>
.cl-wrapper {
  padding: 0 20px 16px;
}

.cl-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.cl-header h4 {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--text-primary);
}

.cl-count {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  background: var(--bg-subtle);
  padding: 1px 6px;
  border-radius: 10px;
}

.cl-list {
  max-height: 240px;
  overflow-y: auto;
}

.cl-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: var(--font-size-sm);
}

.cl-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-add { background: var(--diff-added-text); }
.dot-modify { background: var(--diff-modified-text); }
.dot-delete { background: var(--diff-deleted-text); }
.dot-publish { background: var(--color-primary); }
.dot-default { background: var(--text-placeholder); }

.cl-operator {
  color: var(--text-secondary);
  font-weight: 500;
}

.cl-action {
  font-weight: 500;
}

.act-add { color: var(--diff-added-text); }
.act-modify { color: var(--diff-modified-text); }
.act-delete { color: var(--diff-deleted-text); }
.act-publish { color: var(--color-primary); }
.act-default { color: var(--text-secondary); }

.cl-key {
  font-size: var(--font-size-xs);
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--bg-subtle);
  color: var(--text-secondary);
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.cl-spacer {
  flex: 1;
}

.cl-time {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  white-space: nowrap;
}

.cl-empty {
  text-align: center;
  padding: 20px 0;
  font-size: var(--font-size-sm);
  color: var(--text-placeholder);
}
</style>
