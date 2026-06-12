<template>
  <span class="status-badge" :class="statusClass">
    <span class="sb-dot"></span>
    {{ label }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  hasDraft: { type: Boolean, required: true },
})

const statusClass = computed(() => (props.hasDraft ? 'pending' : 'published'))
const label = computed(() => (props.hasDraft ? 'Pending' : 'Published'))
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 1px 8px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
  white-space: nowrap;
}

.status-badge.published {
  background: var(--color-published-bg);
  color: var(--color-published);
}

.status-badge.pending {
  background: var(--color-pending-bg);
  color: var(--color-pending);
}

.sb-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.published .sb-dot {
  background: var(--color-published);
}

.pending .sb-dot {
  background: var(--color-pending);
}
</style>
