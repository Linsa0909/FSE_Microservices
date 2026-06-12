<template>
  <span class="status-badge" :class="statusClass">
    <span class="dot"></span>
    {{ label }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  hasDraft: { type: Boolean, required: true },
})

const statusClass = computed(() => (props.hasDraft ? 'pending' : 'published'))
const label = computed(() => (props.hasDraft ? '待发布' : '已发布'))
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 10px;
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  font-weight: 500;
  letter-spacing: 0.02em;
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
.dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.published .dot { background: var(--color-published); }
.pending .dot { background: var(--color-pending); }
</style>
