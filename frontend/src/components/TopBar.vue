<template>
  <div class="command-bar">
    <div class="cb-search">
      <svg class="search-icon" width="14" height="14" viewBox="0 0 16 16" fill="none">
        <circle cx="7" cy="7" r="4.5" stroke="#9ca3af" stroke-width="1.5"/>
        <path d="M10.5 10.5L14 14" stroke="#9ca3af" stroke-width="1.5" stroke-linecap="round"/>
      </svg>
      <input :value="search" class="search-input" type="text" placeholder="搜索服务名..."
        @input="$emit('update:search', $event.target.value)" />
    </div>
  </div>
</template>

<script setup>
defineProps({
  search: { type: String, default: '' },
})

defineEmits(['update:search'])
</script>

<style scoped>
.command-bar {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 16px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
}

.cb-search {
  position: relative;
  max-width: 320px;
  width: 100%;
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

.search-input::placeholder { color: var(--text-placeholder); }
.search-input:focus { border-color: var(--color-primary); }
</style>
