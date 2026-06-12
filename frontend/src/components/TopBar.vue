<template>
  <div class="command-bar">
    <div class="cb-left">
      <!-- Search -->
      <div class="cb-search">
        <svg class="search-icon" width="14" height="14" viewBox="0 0 16 16" fill="none">
          <circle cx="7" cy="7" r="4.5" stroke="#9ca3af" stroke-width="1.5"/>
          <path d="M10.5 10.5L14 14" stroke="#9ca3af" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <input :value="search" class="search-input" type="text" placeholder="搜索服务名..."
          @input="$emit('update:search', $event.target.value)" />
      </div>
    </div>

    <div class="cb-right">
      <!-- Env filter -->
      <select :value="env" class="cb-select" @change="$emit('update:env', $event.target.value)">
        <option value="">全部环境</option>
        <option value="dev">dev</option>
        <option value="test">test</option>
        <option value="prod">prod</option>
      </select>

      <!-- Status filter -->
      <select :value="status" class="cb-select" @change="$emit('update:status', $event.target.value)">
        <option value="">全部状态</option>
        <option value="published">已发布</option>
        <option value="pending">待发布</option>
      </select>

      <div class="cb-sep"></div>

      <!-- Refresh info -->
      <span class="cb-refresh-info">
        每 5 秒自动刷新 · <span class="cb-time">{{ lastRefresh }}</span>
      </span>
      <button class="cb-refresh-btn" @click="$emit('manual-refresh')" title="手动刷新">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" :class="{ spinning: refreshing }">
          <path d="M13.65 2.35A7.96 7.96 0 008 0a8 8 0 100 16 7.96 7.96 0 005.65-2.35" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
      </button>

      <!-- Help -->
      <button class="cb-help-btn" @click="showHelp = true" title="帮助">
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6.5" stroke="currentColor" stroke-width="1.3"/>
          <path d="M6.5 6a1.5 1.5 0 012.8-.7M8 9.5V12" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/>
        </svg>
      </button>
    </div>

    <!-- Help dialog -->
    <el-dialog v-model="showHelp" title="操作指南" width="480px">
      <div class="help-content">
        <h4>📋 查看配置</h4>
        <p>左侧边栏选择环境/状态 → 表格显示所有配置组 → <strong>点击任意一行</strong> → 右侧展开详情面板。</p>

        <h4>✏️ 新增配置项</h4>
        <p>展开详情面板 → 在底部 <code>new.key</code> 输入框填写 Key → <code>value</code> 输入框填写值 → 点击 <strong>+</strong> 按钮或按回车。</p>

        <h4>📝 修改配置项</h4>
        <p>在详情面板右侧 <strong>Draft</strong> 列中 <strong>直接点击值</strong> → 变为可编辑输入框 → 修改后按回车保存。</p>

        <h4>🗑️ 删除配置项</h4>
        <p>在详情面板点击每行右侧的 <strong>垃圾桶图标</strong> → 确认后删除。</p>

        <h4>🚀 发布配置</h4>
        <p>编辑完成后 → 点击底部 <strong>发布</strong> 按钮 → 草稿同步到已发布配置 → 版本号 +1。</p>
        <p class="help-note">未做任何修改时发布按钮灰色禁用；有未发布变更时高亮显示。</p>

        <h4>🔄 自动刷新</h4>
        <p>页面每 <strong>5 秒</strong>自动拉取后端最新数据，顶部栏显示上次刷新时间。</p>
        <p>点击右侧 <strong>↻ 刷新按钮</strong> 可手动立即刷新。</p>
      </div>
      <template #footer>
        <el-button @click="showHelp = false">知道了</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  search: { type: String, default: '' },
  env: { type: String, default: '' },
  status: { type: String, default: '' },
  lastRefresh: { type: String, default: '--:--:--' },
  refreshing: { type: Boolean, default: false },
})

defineEmits(['manual-refresh', 'update:search', 'update:env', 'update:status'])

const showHelp = ref(false)
</script>

<style scoped>
.command-bar {
  display: flex; align-items: center; justify-content: space-between;
  height: 44px; padding: 0 16px; background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle); gap: 12px;
}
.cb-left { flex: 1; min-width: 0; }
.cb-search { position: relative; max-width: 320px; }
.search-icon { position: absolute; left: 8px; top: 50%; transform: translateY(-50%); pointer-events: none; }
.search-input {
  width: 100%; height: 28px; padding: 0 8px 0 28px;
  border: 1px solid var(--border-subtle); border-radius: var(--radius-md);
  font-family: var(--font-family); font-size: var(--font-size-sm);
  color: var(--text-primary); background: var(--bg-surface); outline: none;
  transition: border-color 0.15s;
}
.search-input::placeholder { color: var(--text-placeholder); }
.search-input:focus { border-color: var(--color-primary); }
.cb-right { display: flex; align-items: center; gap: 6px; flex-shrink: 0; }
.cb-select {
  height: 28px; padding: 0 24px 0 8px;
  border: 1px solid var(--border-subtle); border-radius: var(--radius-md);
  font-family: var(--font-family); font-size: var(--font-size-xs);
  color: var(--text-secondary); background: var(--bg-surface);
  outline: none; cursor: pointer; appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg width='10' height='10' viewBox='0 0 12 12' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M3 4.5L6 7.5L9 4.5' stroke='%236b7280' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
  background-repeat: no-repeat; background-position: right 6px center;
  transition: border-color 0.15s;
}
.cb-select:focus { border-color: var(--color-primary); }
.cb-sep { width: 1px; height: 20px; background: var(--border-subtle); margin: 0 6px; }
.cb-refresh-info { font-size: var(--font-size-xs); color: var(--text-placeholder); white-space: nowrap; }
.cb-time { font-variant-numeric: tabular-nums; color: var(--text-secondary); }
.cb-refresh-btn, .cb-help-btn {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px;
  border: 1px solid var(--border-subtle); border-radius: var(--radius-md);
  background: var(--bg-surface); color: var(--text-secondary);
  cursor: pointer; transition: all 0.15s;
}
.cb-refresh-btn:hover, .cb-help-btn:hover { border-color: var(--border-hover); color: var(--text-primary); }
.spinning { animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Help content */
.help-content { font-family: var(--font-family); color: var(--text-primary); }
.help-content h4 { font-size: var(--font-size-base); font-weight: 600; margin: 16px 0 4px; color: var(--text-primary); }
.help-content h4:first-child { margin-top: 0; }
.help-content p { font-size: var(--font-size-sm); color: var(--text-secondary); margin: 2px 0 8px; line-height: 1.6; }
.help-content code { font-size: var(--font-size-xs); padding: 1px 5px; border-radius: 3px; background: var(--bg-subtle); border: 1px solid var(--border-subtle); }
.help-note { color: var(--text-placeholder) !important; font-style: italic; }
</style>
