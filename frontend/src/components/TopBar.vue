<template>
  <header class="topbar">
    <div class="topbar-left">
      <svg class="logo" width="22" height="22" viewBox="0 0 22 22" fill="none">
        <rect width="22" height="22" rx="5" fill="#2563EB"/>
        <path d="M6 11l3 3 7-7" stroke="#fff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      <span class="title">配置中心</span>
      <span class="subtitle">Config Center</span>
    </div>
    <div class="topbar-right">
      <span class="refresh-info">
        每 5s 自动刷新 · 上次刷新 <span class="time">{{ lastRefresh }}</span>
      </span>
      <button class="refresh-btn" @click="$emit('manual-refresh')" title="手动刷新">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" :class="{ spinning: refreshing }">
          <path d="M13.65 2.35A7.96 7.96 0 008 0a8 8 0 100 16 7.96 7.96 0 005.65-2.35" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup>
defineProps({
  lastRefresh: { type: String, default: '--:--:--' },
  refreshing: { type: Boolean, default: false },
})
defineEmits(['manual-refresh'])
</script>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 24px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-default);
  position: sticky;
  top: 0;
  z-index: 100;
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo {
  flex-shrink: 0;
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: -0.01em;
}

.subtitle {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  padding-left: 8px;
  border-left: 1px solid var(--border-default);
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.refresh-info {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.refresh-info .time {
  font-variant-numeric: tabular-nums;
  color: var(--color-text);
}

.refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-btn);
  background: var(--bg-card);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: border-color 0.15s, color 0.15s;
}

.refresh-btn:hover {
  border-color: var(--border-hover);
  color: var(--color-text);
}

.spinning {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
