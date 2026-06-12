<template>
  <!-- Desktop: inline side panel -->
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
          <button class="dp-close" @click="$emit('close')" title="关闭">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
          </button>
        </div>
        <div class="dp-header-meta">
          版本 <strong>{{ versionLabel }}</strong>
          <span v-if="config?.lastPublishedAt"> · {{ formatTime(config.lastPublishedAt) }}</span>
        </div>
      </div>

      <!-- Body -->
      <div class="dp-body">
        <!-- Diff summary -->
        <div class="dp-summary">
          <span class="dp-summary-badge" :class="{ clean: diffCount === 0 }">
            {{ diffCount > 0 ? `${diffCount} 项变更待发布` : '草稿与已发布一致' }}
          </span>
          <span class="dp-summary-desc">点击草稿值可修改配置，发布后示例微服务下次拉取 PublishedData。</span>
        </div>

        <!-- Add config section -->
        <section class="add-config-card">
          <div class="section-title">
            <div>
              <h3>新增配置项</h3>
              <p>写入当前配置组的 DraftData，发布前不会影响微服务读取。</p>
            </div>
            <button class="add-submit" @click="addKey" :disabled="!newKey.trim()">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
              添加
            </button>
          </div>
          <div class="add-form">
            <label>
              <span>Key</span>
              <input v-model="newKey" class="dp-inline-input" placeholder="db.url" @keydown.enter="addKey" />
            </label>
            <label>
              <span>Value</span>
              <input v-model="newValue" class="dp-inline-input" placeholder="localhost:3306" @keydown.enter="addKey" />
            </label>
          </div>
        </section>

        <!-- Diff table -->
        <div class="dp-table">
          <div class="dp-th">
            <span class="dp-th-cell key">键</span>
            <span class="dp-th-cell pub">已发布</span>
            <span class="dp-th-cell draft">草稿</span>
            <span class="dp-th-cell act">操作</span>
          </div>

          <div v-for="diff in diffRows" :key="diff.key" class="dp-row" :class="diff.rowClass">
            <div class="dp-cell key"><code>{{ diff.key }}</code></div>
            <div class="dp-cell pub">
              <span :class="{ deleted: diff.deleted }">{{ valueText(diff.publishedVal, diff.added) }}</span>
            </div>
            <div class="dp-cell draft">
              <input
                v-if="editingKey === diff.key"
                v-model="editingValue"
                class="dp-inline-input"
                @keydown.enter="saveInline(diff)"
                @keydown.escape="cancelInline"
                @blur="saveInline(diff)"
              />
              <span
                v-else
                class="dp-editable"
                :class="{ added: diff.added, deleted: diff.deleted }"
                @click="startEdit(diff)"
              >
                {{ valueText(diff.draftVal, diff.deleted) }}
              </span>
            </div>
            <div class="dp-cell act">
              <button class="dp-act-btn" @click="startEdit(diff)" title="编辑">
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M10 1.5l2.5 2.5L4.5 12H2v-2.5L10 1.5z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
              <button class="dp-act-btn danger" @click="confirmDelete(diff.key)" title="删除">
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M2 4h10M5 4V2.5h4V4M3 4v7.5h8V4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- ChangeLog -->
        <div class="dp-section-divider"></div>
        <ChangeLog :logs="logs" />
      </div>

      <!-- Footer -->
      <div class="dp-footer">
        <button class="dp-footer-cancel" @click="$emit('close')">取消</button>
        <button class="dp-footer-publish" :disabled="!hasDraft" @click="doPublish">
          {{ hasDraft ? `发布 (${diffCount} 项变更)` : '已是最新版本' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Mobile: el-drawer -->
  <el-drawer v-else v-model="visible" :title="null" direction="rtl" size="100%" :close-on-click-modal="true" @close="$emit('close')">
    <template #header>
      <div class="dp-title-row">
        <span class="dp-service">{{ config?.service }}</span>
        <span class="dp-sep">/</span>
        <span class="dp-env">{{ config?.env }}</span>
        <StatusBadge :hasDraft="hasDraft" />
      </div>
      <div class="dp-header-meta" style="margin-top:4px">
        版本 <strong>{{ versionLabel }}</strong>
        <span v-if="config?.lastPublishedAt"> · {{ formatTime(config.lastPublishedAt) }}</span>
      </div>
    </template>

    <div style="padding:16px 20px">
      <div class="dp-summary" style="margin-bottom:12px">
        <span class="dp-summary-badge" :class="{ clean: diffCount === 0 }">
          {{ diffCount > 0 ? `${diffCount} 项变更待发布` : '草稿与已发布一致' }}
        </span>
        <span class="dp-summary-desc">点击草稿值可修改配置。</span>
      </div>

      <section class="add-config-card">
        <div class="section-title">
          <div>
            <h3>新增配置项</h3>
            <p>新增内容先进入 DraftData。</p>
          </div>
          <button class="add-submit" @click="addKey" :disabled="!newKey.trim()">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M7 2v10M2 7h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>
            添加
          </button>
        </div>
        <div class="add-form">
          <label>
            <span>Key</span>
            <input v-model="newKey" class="dp-inline-input" placeholder="db.url" @keydown.enter="addKey" />
          </label>
          <label>
            <span>Value</span>
            <input v-model="newValue" class="dp-inline-input" placeholder="localhost:3306" @keydown.enter="addKey" />
          </label>
        </div>
      </section>

      <div class="dp-table">
        <div class="dp-th">
          <span class="dp-th-cell key">键</span>
          <span class="dp-th-cell pub">已发布</span>
          <span class="dp-th-cell draft">草稿</span>
          <span class="dp-th-cell act">操作</span>
        </div>

        <div v-for="diff in diffRows" :key="diff.key" class="dp-row" :class="diff.rowClass">
          <div class="dp-cell key"><code>{{ diff.key }}</code></div>
          <div class="dp-cell pub"><span :class="{ deleted: diff.deleted }">{{ valueText(diff.publishedVal, diff.added) }}</span></div>
          <div class="dp-cell draft">
            <input
              v-if="editingKey === diff.key"
              v-model="editingValue"
              class="dp-inline-input"
              @keydown.enter="saveInline(diff)"
              @keydown.escape="cancelInline"
              @blur="saveInline(diff)"
            />
            <span v-else class="dp-editable" :class="{ added: diff.added, deleted: diff.deleted }" @click="startEdit(diff)">{{ valueText(diff.draftVal, diff.deleted) }}</span>
          </div>
          <div class="dp-cell act">
            <button class="dp-act-btn" @click="startEdit(diff)" title="编辑">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M10 1.5l2.5 2.5L4.5 12H2v-2.5L10 1.5z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
            <button class="dp-act-btn danger" @click="confirmDelete(diff.key)" title="删除">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M2 4h10M5 4V2.5h4V4M3 4v7.5h8V4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
            </button>
          </div>
        </div>
      </div>

      <div class="dp-section-divider"></div>
      <ChangeLog :logs="logs" />
    </div>

    <template #footer>
      <div style="display:flex;justify-content:flex-end;gap:8px">
        <el-button @click="visible=false">取消</el-button>
        <el-button type="primary" :disabled="!hasDraft" @click="doPublish">
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
import ChangeLog from './ChangeLog.vue'
import { hasDraft as calcHasDraft, countDiff, deepEqual } from '../utils/configDiff.js'

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

// BUG-009 fix: component-level editing state instead of on computed row objects
const editingKey = ref('')
const editingValue = ref('')

const hasDraft = computed(() => calcHasDraft(config.value))
const diffCount = computed(() => (config.value ? countDiff(config.value).total : 0))

const versionLabel = computed(() => {
  if (!config.value) return '未加载'
  const v = config.value.publishedVersion
  if (v === undefined || v === null || v === '' || v === 0) return '未发布'
  return `v${v}`
})

const diffRows = computed(() => {
  if (!config.value) return []
  const pub = config.value.publishedData || {}
  const draft = config.value.draftData || {}
  const allKeys = [...new Set([...Object.keys(pub), ...Object.keys(draft)])]

  return allKeys.map(key => {
    const pubVal = pub[key]; const draftVal = draft[key]
    const added = !(key in pub); const deleted = !(key in draft)
    const modified = !added && !deleted && pubVal !== draftVal
    let rowClass = ''
    if (added) rowClass = 'diff-added'
    else if (deleted) rowClass = 'diff-deleted'
    else if (modified) rowClass = 'diff-modified'
    return { key, publishedVal: pubVal, draftVal, originalDraftVal: draftVal, added, deleted, modified, rowClass }
  })
})

function formatTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function valueText(val, added) {
  if (added) return '—'
  if (val === undefined || val === null) return '—'
  return val
}

async function loadDetail() {
  if (!props.service || !props.env) return
  try {
    const data = await getOneConfig(props.service, props.env)
    config.value = data.config; logs.value = data.logs || []
  } catch { ElMessage.error('加载配置详情失败') }
}

watch(
  () => [props.service, props.env, props.modelValue, props.panel],
  ([s, e, v, p]) => { if (s && e && (v || p)) loadDetail() }
)

function startEdit(diff) {
  editingKey.value = diff.key
  editingValue.value = diff.draftVal || ''
  nextTick(() => {
    const inputs = document.querySelectorAll('.dp-row .dp-cell.draft input')
    if (inputs.length) inputs[inputs.length - 1]?.focus()
  })
}

async function saveInline(diff) {
  if (editingKey.value !== diff.key) return
  editingKey.value = ''
  const val = editingValue.value
  if (val === diff.originalDraftVal) return
  try { await setKey(props.service, props.env, diff.key, val); await loadDetail(); ElMessage.success('已保存') }
  catch { ElMessage.error('保存失败') }
}

function cancelInline() { editingKey.value = '' }

async function addKey() {
  const k = newKey.value.trim()
  const v = newValue.value
  if (!k) return
  try { await setKey(props.service, props.env, k, v); newKey.value = ''; newValue.value = ''; await loadDetail(); ElMessage.success('已添加') }
  catch { ElMessage.error('添加失败') }
}

async function confirmDelete(key) {
  try { await ElMessageBox.confirm(`确定删除 "${key}"？`, '确认删除', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }) }
  catch { return }
  try { await deleteKey(props.service, props.env, key); await loadDetail(); ElMessage.success('已删除') }
  catch { ElMessage.error('删除失败') }
}

async function doPublish() {
  try { await publishConfig(props.service, props.env); await loadDetail(); ElMessage.success('发布成功'); emit('published') }
  catch { ElMessage.error('发布失败') }
}
</script>

<style scoped>
.detail-panel { width: min(680px, 44vw); min-width: 400px; height: 100%; background: var(--bg-surface); border-left: 1px solid var(--border-subtle); display: flex; flex-direction: column; overflow: hidden; }
.dp-inner { display: flex; flex-direction: column; height: 100%; overflow: hidden; }
.dp-header { padding: 16px 20px; border-bottom: 1px solid var(--border-subtle); flex-shrink: 0; }
.dp-header-top { display: flex; align-items: center; justify-content: space-between; }
.dp-title-row { display: flex; align-items: center; gap: 8px; }
.dp-service { font-size: var(--font-size-lg); font-weight: 600; color: var(--text-primary); }
.dp-sep { color: var(--text-placeholder); font-weight: 300; }
.dp-env { font-size: var(--font-size-base); font-weight: 500; color: var(--text-secondary); }
.dp-close { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; border: 1px solid var(--border-subtle); border-radius: var(--radius-md); background: var(--bg-surface); color: var(--text-secondary); cursor: pointer; transition: all 0.15s; }
.dp-close:hover { border-color: var(--border-hover); color: var(--text-primary); }
.dp-header-meta { margin-top: 6px; font-size: var(--font-size-xs); color: var(--text-placeholder); }
.dp-header-meta strong { color: var(--text-secondary); font-weight: 500; }
.dp-body { flex: 1; overflow-y: auto; padding: 16px 20px; }

/* Summary */
.dp-summary { margin-bottom: 12px; }
.dp-summary-badge { display: inline-block; padding: 3px 10px; border-radius: var(--radius-sm); background: var(--color-pending-bg); color: var(--color-pending); font-size: var(--font-size-xs); font-weight: 500; }
.dp-summary-badge.clean { background: var(--color-published-bg); color: var(--color-published); }
.dp-summary-desc { display: block; margin-top: 4px; font-size: var(--font-size-xs); color: var(--text-placeholder); }

/* Add config card */
.add-config-card { padding: 12px; border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); margin-bottom: 14px; }
.section-title { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 10px; }
.section-title h3 { font-size: var(--font-size-sm); font-weight: 600; color: var(--text-primary); }
.section-title p { font-size: var(--font-size-xs); color: var(--text-placeholder); margin-top: 2px; }
.add-submit { display: inline-flex; align-items: center; gap: 5px; height: 28px; padding: 0 10px; border: 1px solid var(--color-primary); border-radius: var(--radius-md); background: var(--color-primary); color: #fff; font-family: var(--font-family); font-size: var(--font-size-xs); font-weight: 500; cursor: pointer; flex-shrink: 0; }
.add-submit:hover { background: var(--color-primary-hover); }
.add-submit:disabled { opacity: 0.4; cursor: not-allowed; }
.add-form { display: flex; gap: 8px; }
.add-form label { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.add-form label span { font-size: var(--font-size-xs); color: var(--text-secondary); font-weight: 500; }

/* Diff table */
.dp-table { border: 1px solid var(--border-subtle); border-radius: var(--radius-md); overflow: hidden; }
.dp-th { display: flex; align-items: center; height: 30px; padding: 0 12px; border-bottom: 1px solid var(--border-subtle); background: var(--bg-subtle); }
.dp-th-cell { font-size: var(--font-size-xs); color: var(--text-placeholder); font-weight: 500; }
.key { width: 30%; } .pub { width: 28%; } .draft { width: 28%; } .act { width: 14%; text-align: right; }
.dp-row { display: flex; align-items: center; min-height: 38px; padding: 4px 12px; border-bottom: 1px solid var(--border-subtle); transition: background 0.1s; }
.dp-row:last-child { border-bottom: none; }
.dp-row:hover { background: var(--bg-hover); }
.dp-row.diff-added { background: var(--diff-added-bg); border-left: 3px solid var(--diff-added-text); }
.dp-row.diff-modified { background: var(--diff-modified-bg); border-left: 3px solid var(--diff-modified-text); }
.dp-row.diff-deleted { background: var(--diff-deleted-bg); border-left: 3px solid var(--diff-deleted-text); }
.dp-cell { font-size: var(--font-size-sm); color: var(--text-primary); overflow: hidden; }
.dp-cell code { font-size: var(--font-size-xs); padding: 2px 6px; border-radius: 3px; background: var(--bg-subtle); border: 1px solid var(--border-subtle); word-break: break-all; }
.added { color: var(--diff-added-text); } .deleted { color: var(--diff-deleted-text); text-decoration: line-through; }
.dp-editable { cursor: pointer; padding: 2px 6px; border-radius: var(--radius-sm); border: 1px solid transparent; transition: border-color 0.15s, background 0.15s; display: inline-block; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dp-editable:hover { border-color: var(--color-primary); background: var(--bg-blue); }
.dp-inline-input { width: 100%; height: 26px; padding: 0 6px; border: 1px solid var(--color-primary); border-radius: var(--radius-sm); font-family: var(--font-family); font-size: var(--font-size-sm); outline: none; }
.dp-act-btn { display: inline-flex; align-items: center; justify-content: center; width: 26px; height: 26px; border: 1px solid var(--border-subtle); border-radius: var(--radius-sm); background: var(--bg-surface); color: var(--text-secondary); cursor: pointer; transition: all 0.15s; margin-left: 4px; }
.dp-act-btn:hover { border-color: var(--border-hover); color: var(--text-primary); }
.dp-act-btn.primary:hover { border-color: var(--color-primary); color: var(--color-primary); }
.dp-act-btn.danger:hover { border-color: var(--diff-deleted-text); color: var(--diff-deleted-text); }
.dp-act-btn:disabled { opacity: 0.35; cursor: not-allowed; }

/* Section divider */
.dp-section-divider { height: 1px; background: var(--border-subtle); margin: 16px 0 12px; }

/* Footer */
.dp-footer { display: flex; justify-content: flex-end; gap: 8px; padding: 12px 20px; border-top: 1px solid var(--border-subtle); flex-shrink: 0; }
.dp-footer-cancel { height: 32px; padding: 0 12px; border: 1px solid var(--border-subtle); border-radius: var(--radius-md); background: var(--bg-surface); color: var(--text-secondary); font-family: var(--font-family); font-size: var(--font-size-sm); font-weight: 500; cursor: pointer; transition: all 0.15s; }
.dp-footer-cancel:hover { border-color: var(--border-hover); color: var(--text-primary); }
.dp-footer-publish { height: 32px; padding: 0 14px; border: none; border-radius: var(--radius-md); background: var(--color-primary); color: #fff; font-family: var(--font-family); font-size: var(--font-size-sm); font-weight: 500; cursor: pointer; transition: background 0.15s; }
.dp-footer-publish:hover { background: var(--color-primary-hover); }
.dp-footer-publish:disabled { opacity: 0.4; cursor: not-allowed; }
</style>
