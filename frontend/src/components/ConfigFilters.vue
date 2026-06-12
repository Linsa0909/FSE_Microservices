<template>
  <div class="filters">
    <div class="filter-search">
      <svg class="search-icon" width="16" height="16" viewBox="0 0 16 16" fill="none">
        <circle cx="7" cy="7" r="4.5" stroke="#9CA3AF" stroke-width="1.5"/>
        <path d="M10.5 10.5L14 14" stroke="#9CA3AF" stroke-width="1.5" stroke-linecap="round"/>
      </svg>
      <input
        v-model="searchText"
        class="search-input"
        type="text"
        placeholder="搜索服务名..."
        @input="emitFilters"
      />
    </div>
    <div class="filter-selects">
      <select v-model="envFilter" class="filter-select" @change="emitFilters">
        <option value="">全部环境</option>
        <option value="dev">dev</option>
        <option value="test">test</option>
        <option value="prod">prod</option>
      </select>
      <select v-model="statusFilter" class="filter-select" @change="emitFilters">
        <option value="">全部状态</option>
        <option value="published">PUBLISHED</option>
        <option value="pending">PENDING_PUBLISH</option>
      </select>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['filter'])

const searchText = ref('')
const envFilter = ref('')
const statusFilter = ref('')

function emitFilters() {
  emit('filter', {
    search: searchText.value.toLowerCase(),
    env: envFilter.value,
    status: statusFilter.value,
  })
}
</script>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 24px;
}

.filter-search {
  position: relative;
  flex: 1;
  max-width: 320px;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 36px;
  padding: 0 12px 0 32px;
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-input);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  color: var(--color-text);
  background: var(--bg-card);
  outline: none;
  transition: border-color 0.15s;
}

.search-input::placeholder {
  color: var(--color-text-placeholder);
}

.search-input:focus {
  border-color: var(--color-primary);
}

.filter-selects {
  display: flex;
  gap: 8px;
}

.filter-select {
  height: 36px;
  padding: 0 28px 0 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-input);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  color: var(--color-text);
  background: var(--bg-card);
  outline: none;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='12' height='12' viewBox='0 0 12 12' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M3 4.5L6 7.5L9 4.5' stroke='%236B7280' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  transition: border-color 0.15s;
}

.filter-select:focus {
  border-color: var(--color-primary);
}
</style>
