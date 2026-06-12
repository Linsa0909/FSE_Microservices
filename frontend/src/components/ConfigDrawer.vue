<template>
  <!-- Desktop: inline panel -->
  <div v-if="panel" class="detail-panel">
    <div class="dp-inner">
      <!-- Header -->
      <div class="dp-header">
        <div class="dp-header-top">
          <div class="dp-title-row">
            <span class="dp-service">{{ config?.service }}</span>
            <span class="dp-sep">/</span>
            <span class="dp-env">{{ config?.env }}</span>
            <StatusBadge :hasDraft="hasDraft" />
          </div>
          <button class="dp-close" @click="$emit('close')" title="Close">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
        <div class="dp-header-meta">
          Version <strong>v{{ config?.publishedVersion }}</strong>
          <span v-if="config?.lastPublishedAt"> · {{ formatTime(config.lastPublishedAt) }}</span>
        </div>
      </div>

      <!-- Body -->
      <div class="dp-body">
        <DetailBody
          :diffRows="diffRows"
          :diffCount="diffCount"
          :logs="logs"
          :newKey="newKey"
          :newValue="newValue"
          @add-key="addKey"
          @save-inline="saveInline"
          @cancel-inline="cancelInline"
          @start-edit="startEdit"
          @confirm-delete="confirmDelete"
          @update:newKey="newKey = $event"
          @update:newValue="newValue = $event"
        />
      </div>

      <!-- Footer -->
      <div class="dp-footer">
        <button class="dp-footer-cancel" @click="$emit('close')">Cancel</button>
        <button
          class="dp-footer-publish"
          :disabled="!hasDraft"
          @click="doPublish"
        >
          {{ hasDraft ? `Publish (${diffCount} changes)` : 'Up to date' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Mobile: el-drawer -->
  <el-drawer
    v-else
    v-model="visible"
    :title="null"
    direction="rtl"
    size="100%"
    :close-on-click-modal="true"
    @close="$emit('close')"
  >
    <!-- Header -->
    <div class="dp-header">
      <div class="dp-header-top">
        <div class="dp-title-row">
          <span class="dp-service">{{ config?.service }}</span>
          <span class="dp-sep">/</span>
          <span class="dp-env">{{ config?.env }}</span>
          <StatusBadge :hasDraft="hasDraft" />
        </div>
      </div>
      <div class="dp-header-meta">
        Version <strong>v{{ config?.publishedVersion }}</strong>
        <span v-if="config?.lastPublishedAt"> · {{ formatTime(config.lastPublishedAt) }}</span>
      </div>
    </div>

    <!-- Body -->
    <div class="dp-body">
      <DetailBody
        :diffRows="diffRows"
        :diffCount="diffCount"
        :logs="logs"
        :newKey="newKey"
        :newValue="newValue"
        @add-key="addKey"
        @save-inline="saveInline"
        @cancel-inline="cancelInline"
        @start-edit="startEdit"
        @confirm-delete="confirmDelete"
        @update:newKey="newKey = $event"
        @update:newValue="newValue = $event"
      />
    </div>

    <template #footer>
      <div class="drawer-footer">
        <el-button @click="visible = false">Cancel</el-button>
        <el-button type="primary" :disabled="!hasDraft" @click="doPublish">
          {{ hasDraft ? `Publish (${diffCount} changes)` : 'Up to date' }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<!-- ============================================================ -->
<!-- DetailBody — shared content between panel and drawer          -->
<!-- ============================================================ -->
<script>
import { h } from 'vue'

const DetailBody = {
  name: 'DetailBody',
  props: {
    diffRows: Array,
    diffCount: Number,
    logs: Array,
    newKey: String,
    newValue: String,
  },
  emits: ['add-key', 'save-inline', 'cancel-inline', 'start-edit', 'confirm-delete'],
  setup(props, { emit }) {
    function onAddKey() { emit('add-key') }
    function onSaveInline(diff) { emit('save-inline', diff) }
    function onCancelInline(diff) { emit('cancel-inline', diff) }
    function onStartEdit(diff) { emit('start-edit', diff) }
    function onConfirmDelete(key) { emit('confirm-delete', key) }
    return { onAddKey, onSaveInline, onCancelInline, onStartEdit, onConfirmDelete }
  },
  template: `
    <div>
      <!-- Diff summary -->
      <div class="dp-summary" v-if="diffCount > 0">
        <span class="dp-summary-badge">{{ diffCount }} change{{ diffCount > 1 ? 's' : '' }} pending</span>
      </div>

      <!-- Diff table -->
      <div class="dp-table">
        <div class="dp-th">
          <span class="dp-th-cell key">Key</span>
          <span class="dp-th-cell pub">Published</span>
          <span class="dp-th-cell draft">Draft</span>
          <span class="dp-th-cell act">Actions</span>
        </div>

        <div v-for="diff in diffRows" :key="diff.key" class="dp-row" :class="diff.rowClass">
          <div class="dp-cell key">
            <code>{{ diff.key }}</code>
          </div>
          <div class="dp-cell pub">
            <span :class="{ deleted: diff.deleted }">{{ diff.publishedVal || '—' }}</span>
          </div>
          <div class="dp-cell draft">
            <input
              v-if="diff.editing"
              v-model="diff.editValue"
              class="dp-inline-input"
              @keydown.enter="onSaveInline(diff)"
              @keydown.escape="onCancelInline(diff)"
              @blur="onSaveInline(diff)"
            />
            <span
              v-else
              class="dp-editable"
              :class="{ added: diff.added, deleted: diff.deleted }"
              @click="onStartEdit(diff)"
            >
              {{ diff.draftVal || '—' }}
            </span>
          </div>
          <div class="dp-cell act">
            <button class="dp-act-btn" @click="onStartEdit(diff)" title="Edit">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M10 1.5l2.5 2.5L4.5 12H2v-2.5L10 1.5z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
            <button class="dp-act-btn danger" @click="onConfirmDelete(diff.key)" title="Delete">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M2 4h10M5 4V2.5h4V4M3 4v7.5h8V4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Add row -->
        <div class="dp-row add-row">
          <div class="dp-cell key">
            <input
              :value="newKey"
              class="dp-inline-input"
              placeholder="new.key"
              @keydown.enter="onAddKey()"
              @input="$emit('update:newKey', $event.target.value)"
            />
          </div>
          <div class="dp-cell pub"><span class="muted">—</span></div>
          <div class="dp-cell draft">
            <input
              :value="newValue"
              class="dp-inline-input"
              placeholder="value"
              @keydown.enter="onAddKey()"
              @input="$emit('update:newValue', $event.target.value)"
            />
          </div>
          <div class="dp-cell act">
            <button class="dp-act-btn primary" @click="onAddKey()" title="Add" :disabled="!newKey">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- ChangeLog -->
      <ChangeLog :logs="logs" />
    </div>
  `
}
</script>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOneConfig, setKey, deleteKey, publishConfig } from '../api/configService.js'
import StatusBadge from './StatusBadge.vue'
import ChangeLog from './ChangeLog.vue'
import { hasDraft as calcHasDraft, countDiff } from '../utils/configDiff.js'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  service: { type: String, default: '' },
  env: { type: String, default: '' },
  panel: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'close', 'published'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const config = ref(null)
const logs = ref([])
const newKey = ref('')
const newValue = ref('')

const hasDraft = computed(() => calcHasDraft(config.value))
const diffCount = computed(() => countDiff(config.value).total)

const diffRows = computed(() => {
  if (!config.value) return []
  const pub = config.value.publishedData || {}
  const draft = config.value.draftData || {}
  const allKeys = new Set([...Object.keys(pub), ...Object.keys(draft)])

  return Array.from(allKeys).map(key => {
    const pubVal = pub[key]
    const draftVal = draft[key]
    const added = !(key in pub)
    const deleted = !(key in draft)
    const modified = !added && !deleted && pubVal !== draftVal

    let rowClass = ''
    if (added) rowClass = 'diff-added'
    else if (deleted) rowClass = 'diff-deleted'
    else if (modified) rowClass = 'diff-modified'

    return {
      key,
      publishedVal: pubVal,
      draftVal,
      originalDraftVal: draftVal,
      added,
      deleted,
      modified,
      rowClass,
      editing: false,
      editValue: draftVal || '',
    }
  })
})

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  return d.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

async function loadDetail() {
  if (!props.service || !props.env) return
  try {
    const data = await getOneConfig(props.service, props.env)
    config.value = data.config
    logs.value = data.logs || []
  } catch (e) {
    ElMessage.error('Failed to load config detail')
  }
}

// Watch for opening — handles both drawer (modelValue) and panel (panel + service/env)
watch(
  () => [props.service, props.env, props.modelValue, props.panel],
  ([s, e, v, p]) => {
    if (s && e && (v || p)) loadDetail()
  }
)

function startEdit(diff) {
  diff.editValue = diff.draftVal || ''
  diff.editing = true
  nextTick(() => {
    const inputs = document.querySelectorAll('.dp-row .dp-cell.draft input')
    if (inputs.length) inputs[inputs.length - 1]?.focus()
  })
}

async function saveInline(diff) {
  if (!diff.editing) return
  diff.editing = false
  const val = (diff.editValue || '').trim()
  if (val === diff.originalDraftVal) return
  try {
    await setKey(props.service, props.env, diff.key, val)
    diff.draftVal = val
    diff.originalDraftVal = val
    await loadDetail()
    ElMessage.success('Saved')
  } catch (e) {
    ElMessage.error('Save failed')
  }
}

function cancelInline(diff) {
  diff.editing = false
  diff.editValue = diff.originalDraftVal
}

async function addKey() {
  const key = newKey.value.trim()
  const val = newValue.value.trim()
  if (!key) return
  try {
    await setKey(props.service, props.env, key, val)
    newKey.value = ''
    newValue.value = ''
    await loadDetail()
    ElMessage.success('Added')
  } catch (e) {
    ElMessage.error('Add failed')
  }
}

async function confirmDelete(key) {
  try {
    await ElMessageBox.confirm(`Delete "${key}"?`, 'Confirm delete', {
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteKey(props.service, props.env, key)
    await loadDetail()
    ElMessage.success('Deleted')
  } catch (e) {
    ElMessage.error('Delete failed')
  }
}

async function doPublish() {
  try {
    await publishConfig(props.service, props.env)
    await loadDetail()
    ElMessage.success('Published')
    emit('published')
  } catch (e) {
    ElMessage.error('Publish failed')
  }
}
</script>

<style scoped>
/* === Detail Panel (desktop) === */
.detail-panel {
  width: 680px;
  min-width: 680px;
  height: 100%;
  background: var(--bg-surface);
  border-left: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dp-inner {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* === Header === */
.dp-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.dp-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dp-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dp-service {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
}

.dp-sep {
  color: var(--text-placeholder);
  font-weight: 300;
}

.dp-env {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--text-secondary);
}

.dp-close {
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

.dp-close:hover {
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.dp-header-meta {
  margin-top: 6px;
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
}

.dp-header-meta strong {
  color: var(--text-secondary);
  font-weight: 500;
}

/* === Body === */
.dp-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

/* Summary */
.dp-summary {
  margin-bottom: 12px;
}

.dp-summary-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: var(--radius-sm);
  background: var(--color-pending-bg);
  color: var(--color-pending);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

/* Diff table */
.dp-table {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.dp-th {
  display: flex;
  align-items: center;
  height: 30px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-subtle);
}

.dp-th-cell {
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  font-weight: 500;
}

.key { width: 30%; }
.pub { width: 28%; }
.draft { width: 28%; }
.act { width: 14%; text-align: right; }

.dp-row {
  display: flex;
  align-items: center;
  min-height: 38px;
  padding: 4px 12px;
  border-bottom: 1px solid var(--border-subtle);
  transition: background 0.1s;
}

.dp-row:last-child { border-bottom: none; }
.dp-row:hover { background: var(--bg-hover); }

/* Diff highlights */
.dp-row.diff-added { background: var(--diff-added-bg); border-left: 3px solid var(--diff-added-text); }
.dp-row.diff-modified { background: var(--diff-modified-bg); border-left: 3px solid var(--diff-modified-text); }
.dp-row.diff-deleted { background: var(--diff-deleted-bg); border-left: 3px solid var(--diff-deleted-text); }
.dp-row.add-row { background: var(--bg-subtle); border-left: 3px solid transparent; }

.dp-cell {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  overflow: hidden;
}

.dp-cell code {
  font-size: var(--font-size-xs);
  padding: 2px 6px;
  border-radius: 3px;
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
  word-break: break-all;
}

.added { color: var(--diff-added-text); }
.deleted { color: var(--diff-deleted-text); text-decoration: line-through; }
.muted { color: var(--text-placeholder); }

.dp-editable {
  cursor: pointer;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: border-color 0.15s, background 0.15s;
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dp-editable:hover {
  border-color: var(--color-primary);
  background: var(--bg-blue);
}

.dp-inline-input {
  width: 100%;
  height: 26px;
  padding: 0 6px;
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-sm);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  outline: none;
}

.dp-act-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-surface);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  margin-left: 4px;
}

.dp-act-btn:hover { border-color: var(--border-hover); color: var(--text-primary); }
.dp-act-btn.primary:hover { border-color: var(--color-primary); color: var(--color-primary); }
.dp-act-btn.danger:hover { border-color: var(--diff-deleted-text); color: var(--diff-deleted-text); }
.dp-act-btn:disabled { opacity: 0.35; cursor: not-allowed; }

/* === Footer (panel mode) === */
.dp-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.dp-footer-cancel {
  height: 32px;
  padding: 0 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  color: var(--text-secondary);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.dp-footer-cancel:hover {
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.dp-footer-publish {
  height: 32px;
  padding: 0 14px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--color-primary);
  color: #fff;
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.dp-footer-publish:hover {
  background: var(--color-primary-hover);
}

.dp-footer-publish:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Drawer footer */
.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
