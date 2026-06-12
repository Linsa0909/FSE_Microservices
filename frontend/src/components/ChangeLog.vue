<template>
  <div class="changelog" v-if="logs && logs.length > 0">
    <div class="changelog-header">
      <h4>变更记录</h4>
      <span class="count">{{ logs.length }} 条</span>
    </div>
    <div class="log-list">
      <div v-for="(log, i) in logs" :key="i" class="log-item">
        <span class="log-time">{{ formatTime(log.time) }}</span>
        <span class="log-operator">{{ log.operator }}</span>
        <span class="log-action" :class="'action-' + actionClass(log.action)">{{ log.action }}</span>
        <code v-if="log.key" class="log-key">{{ log.key }}</code>
        <span class="log-version">v{{ log.version }}</span>
      </div>
    </div>
  </div>
  <div v-else class="changelog-empty">
    <span class="muted">暂无变更记录</span>
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
.changelog {
  margin-top: 24px;
  border-top: 1px solid var(--border-default);
  padding-top: 20px;
}

.changelog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.changelog-header h4 {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text);
}

.count {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.log-list {
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-input);
  overflow: hidden;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 36px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-default);
  font-size: var(--font-size-sm);
}

.log-item:last-child { border-bottom: none; }

.log-time { color: var(--color-text-placeholder); white-space: nowrap; }
.log-operator { color: var(--color-text-secondary); font-weight: 500; }
.log-action { font-weight: 500; }

.action-add { color: var(--diff-added-text); }
.action-modify { color: var(--diff-modified-text); }
.action-delete { color: var(--diff-deleted-text); }
.action-publish { color: var(--color-primary); }
.action-default { color: var(--color-text-secondary); }

.log-key {
  font-size: var(--font-size-xs);
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--bg-page);
}

.log-version { color: var(--color-text-placeholder); margin-left: auto; }

.changelog-empty {
  margin-top: 24px;
  border-top: 1px solid var(--border-default);
  padding-top: 20px;
  text-align: center;
}

.muted { color: var(--color-text-placeholder); font-size: var(--font-size-sm); }
</style>
