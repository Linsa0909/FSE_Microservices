<template>
  <el-drawer
    v-model="visible"
    :title="null"
    direction="rtl"
    size="600px"
    :close-on-click-modal="true"
    @close="$emit('close')"
  >
    <template #header>
      <div class="drawer-header">
        <div class="drawer-title-row">
          <span class="drawer-service">{{ config?.service }}</span>
          <span class="drawer-separator">/</span>
          <span class="drawer-env">{{ config?.env }}</span>
          <StatusBadge :hasDraft="hasDraft" />
        </div>
        <div class="drawer-version">
          Published Version <strong>v{{ config?.publishedVersion }}</strong>
          <span v-if="config?.lastPublishedAt" class="drawer-time">
            · {{ formatTime(config.lastPublishedAt) }}
          </span>
        </div>
      </div>
    </template>

    <div class="drawer-body">
      <!-- Diff summary -->
      <div class="diff-summary" v-if="diffCount > 0">
        <span class="diff-badge">{{ diffCount }} 项变更待发布</span>
      </div>

      <!-- Config keys table -->
      <div class="config-keys-table">
        <div class="key-header">
          <span class="kh kh-key">Key</span>
          <span class="kh kh-pub">Published Value</span>
          <span class="kh kh-draft">Draft Value</span>
          <span class="kh kh-action">Action</span>
        </div>

        <div v-for="diff in diffRows" :key="diff.key" class="key-row" :class="diff.rowClass">
          <div class="kd kd-key">
            <code>{{ diff.key }}</code>
          </div>
          <div class="kd kd-pub">
            <span :class="{ deleted: diff.deleted }">{{ diff.publishedVal || '—' }}</span>
          </div>
          <div class="kd kd-draft">
            <input
              v-if="diff.editing"
              ref="editInputs"
              v-model="diff.editValue"
              class="inline-input"
              @keydown.enter="saveInline(diff)"
              @keydown.escape="cancelInline(diff)"
              @blur="saveInline(diff)"
            />
            <span
              v-else
              class="editable-value"
              :class="{ added: diff.added, deleted: diff.deleted }"
              @click="startEdit(diff)"
            >
              {{ diff.draftVal || '—' }}
            </span>
          </div>
          <div class="kd kd-action">
            <button class="action-btn" @click="startEdit(diff)" title="编辑">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M10 1.5l2.5 2.5L4.5 12H2v-2.5L10 1.5z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
            <button class="action-btn danger" @click="confirmDelete(diff.key)" title="删除">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M2 4h10M5 4V2.5h4V4M3 4v7.5h8V4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Add new key row -->
        <div class="key-row add-row">
          <div class="kd kd-key">
            <input
              v-model="newKey"
              class="inline-input"
              placeholder="new.key"
              @keydown.enter="addKey"
            />
          </div>
          <div class="kd kd-pub">
            <span class="muted">—</span>
          </div>
          <div class="kd kd-draft">
            <input
              v-model="newValue"
              class="inline-input"
              placeholder="value"
              @keydown.enter="addKey"
            />
          </div>
          <div class="kd kd-action">
            <button class="action-btn primary" @click="addKey" title="添加" :disabled="!newKey">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Change log (placeholder) -->
      <!-- <ChangeLog :logs="logs" /> -->
    </div>

    <!-- Footer actions -->
    <template #footer>
      <div class="drawer-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!hasDraft"
          @click="doPublish"
        >
          {{ hasDraft ? `发布 (${diffCount} 项变更)` : '已是最新版本' }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOneConfig, setKey, deleteKey, publishConfig } from '../api/configService.js'
import StatusBadge from './StatusBadge.vue'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  service: { type: String, default: '' },
  env: { type: String, default: '' },
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
const loading = ref(false)

const hasDraft = computed(() => {
  if (!config.value) return false
  return JSON.stringify(config.value.draftData) !== JSON.stringify(config.value.publishedData)
})

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

const diffCount = computed(() => diffRows.value.filter(d => d.added || d.modified || d.deleted).length)

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  return d.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

async function loadDetail() {
  if (!props.service || !props.env) return
  loading.value = true
  try {
    const data = await getOneConfig(props.service, props.env)
    config.value = data.config
    logs.value = data.logs || []
  } catch (e) {
    ElMessage.error('加载配置详情失败')
  } finally {
    loading.value = false
  }
}

watch(() => [props.service, props.env, props.modelValue], ([s, e, v]) => {
  if (v && s && e) loadDetail()
})

function startEdit(diff) {
  diff.editValue = diff.draftVal || ''
  diff.editing = true
  nextTick(() => {
    const inputs = document.querySelectorAll('.key-row .kd-draft input')
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
    // Refresh full config
    await loadDetail()
    ElMessage.success('已保存')
  } catch (e) {
    ElMessage.error('保存失败')
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
    ElMessage.success('已添加')
  } catch (e) {
    ElMessage.error('添加失败')
  }
}

async function confirmDelete(key) {
  try {
    await ElMessageBox.confirm(`确定要删除配置项 "${key}"？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteKey(props.service, props.env, key)
    await loadDetail()
    ElMessage.success('已删除')
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

async function doPublish() {
  try {
    await publishConfig(props.service, props.env)
    await loadDetail()
    ElMessage.success('发布成功')
    emit('published')
  } catch (e) {
    ElMessage.error('发布失败')
  }
}
</script>

<style scoped>
.drawer-header {
  padding-right: 32px;
}

.drawer-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.drawer-service {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: -0.01em;
}

.drawer-separator {
  color: var(--color-text-placeholder);
  font-weight: 300;
}

.drawer-env {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.drawer-version {
  margin-top: 6px;
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.drawer-version strong {
  color: var(--color-text);
}

.drawer-time {
  color: var(--color-text-placeholder);
}

.drawer-body {
  padding: 0;
}

.diff-summary {
  padding: 10px 0;
}

.diff-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: var(--border-radius-tag);
  background: var(--color-pending-bg);
  color: var(--color-pending);
  font-size: var(--font-size-xs);
  font-weight: 500;
}

/* Config keys table */
.config-keys-table {
  margin-top: 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--border-radius-input);
  overflow: hidden;
}

.key-header {
  display: flex;
  align-items: center;
  height: 34px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-page);
}

.kh {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
  font-weight: 500;
}

.kh-key { width: 28%; }
.kh-pub { width: 28%; }
.kh-draft { width: 28%; }
.kh-action { width: 16%; text-align: right; }

.key-row {
  display: flex;
  align-items: center;
  min-height: 40px;
  padding: 6px 12px;
  border-bottom: 1px solid var(--border-default);
  transition: background 0.1s;
}

.key-row:last-child {
  border-bottom: none;
}

.key-row:hover {
  background: var(--bg-hover);
}

/* Diff highlight rows */
.key-row.diff-added { background: var(--diff-added-bg); border-left: 3px solid var(--diff-added-text); }
.key-row.diff-modified { background: var(--diff-modified-bg); border-left: 3px solid var(--diff-modified-text); }
.key-row.diff-deleted { background: var(--diff-deleted-bg); border-left: 3px solid var(--diff-deleted-text); }
.key-row.add-row { background: var(--bg-page); border-left: 3px solid transparent; }

.kd {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  overflow: hidden;
}

.kd-key { width: 28%; }
.kd-pub { width: 28%; }
.kd-draft { width: 28%; }
.kd-action { width: 16%; text-align: right; }

.kd-key code {
  font-size: var(--font-size-xs);
  padding: 2px 6px;
  border-radius: 3px;
  background: var(--bg-page);
  border: 1px solid var(--border-default);
  word-break: break-all;
}

.added { color: var(--diff-added-text); }
.deleted { color: var(--diff-deleted-text); text-decoration: line-through; }
.muted { color: var(--color-text-placeholder); }

.editable-value {
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid transparent;
  transition: border-color 0.15s, background 0.15s;
  display: inline-block;
}

.editable-value:hover {
  border-color: var(--color-primary);
  background: var(--bg-blue);
}

.inline-input {
  width: 100%;
  height: 28px;
  padding: 0 6px;
  border: 1px solid var(--color-primary);
  border-radius: 4px;
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  outline: none;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-default);
  border-radius: 4px;
  background: var(--bg-card);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  margin-left: 4px;
}

.action-btn:hover { border-color: var(--border-hover); color: var(--color-text); }
.action-btn.primary:hover { border-color: var(--color-primary); color: var(--color-primary); }
.action-btn.danger:hover { border-color: var(--diff-deleted-text); color: var(--diff-deleted-text); }
.action-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Footer */
.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
