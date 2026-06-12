<template>
  <div class="command-bar">
    <div class="cb-left">
      <!-- Search -->
      <div class="cb-search">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 16 16" fill="none">
          <circle cx="7" cy="7" r="4.5" stroke="#9ca3af" stroke-width="1.5"/>
          <path d="M10.5 10.5L14 14" stroke="#9ca3af" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <input
          :value="search"
          class="search-input"
          type="text"
          placeholder="Search services..."
          @input="$emit('update:search', $event.target.value)"
        />
      </div>
    </div>

    <div class="cb-right">
      <!-- Env filter -->
      <select :value="env" class="cb-select" @change="$emit('update:env', $event.target.value)">
        <option value="">All envs</option>
        <option value="dev">dev</option>
        <option value="test">test</option>
        <option value="prod">prod</option>
      </select>

      <!-- Status filter -->
      <select :value="status" class="cb-select" @change="$emit('update:status', $event.target.value)">
        <option value="">All status</option>
        <option value="published">Published</option>
        <option value="pending">Pending publish</option>
      </select>

      <div class="cb-sep"></div>

      <!-- Refresh info -->
      <span class="cb-refresh-info">
        Every 5s · <span class="cb-time">{{ lastRefresh }}</span>
      </span>
      <button class="cb-refresh-btn" @click="$emit('manual-refresh')" title="Refresh now">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" :class="{ spinning: refreshing }">
          <path d="M13.65 2.35A7.96 7.96 0 008 0a8 8 0 100 16 7.96 7.96 0 005.65-2.35" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  search: { type: String, default: '' },
  env: { type: String, default: '' },
  status: { type: String, default: '' },
  lastRefresh: { type: String, default: '--:--:--' },
  refreshing: { type: Boolean, default: false },
})

defineEmits(['manual-refresh', 'update:search', 'update:env', 'update:status'])
</script>

<style scoped>
.command-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 44px;
  padding: 0 16px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
  gap: 12px;
}

.cb-left {
  flex: 1;
  min-width: 0;
}

.cb-search {
  position: relative;
  max-width: 320px;
}

.search-icon {
  position: absolute;
  left: 8px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 28px;
  padding: 0 8px 0 28px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  background: var(--bg-surface);
  outline: none;
  transition: border-color 0.15s;
}

.search-input::placeholder {
  color: var(--text-placeholder);
}

.search-input:focus {
  border-color: var(--color-primary);
}

.cb-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.cb-select {
  height: 28px;
  padding: 0 24px 0 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-family: var(--font-family);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  background: var(--bg-surface);
  outline: none;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='10' viewBox='0 0 12 12' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M3 4.5L6 7.5L9 4.5' stroke='%236b7280' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 6px center;
  transition: border-color 0.15s;
}

.cb-select:focus {
  border-color: var(--color-primary);
}

.cb-sep {
  width: 1px;
  height: 20px;
  background: var(--border-subtle);
  margin: 0 6px;
}

.cb-refresh-info {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  white-space: nowrap;
}

.cb-time {
  font-variant-numeric: tabular-nums;
  color: var(--text-secondary);
}

.cb-refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}

.cb-refresh-btn:hover {
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.spinning {
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
