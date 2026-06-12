<template>
  <section class="service-status">
    <div class="panel-head">
      <div>
        <h2>微服务状态</h2>
        <p>实时检测各领域微服务 /health 端点，验证配置拉取与运行状态。</p>
      </div>
      <button class="panel-action" @click="checkAll">
        <el-icon><Refresh /></el-icon>
        刷新全部
      </button>
    </div>

    <div class="service-scroll">
      <article
        v-for="svc in serviceViews"
        :key="`${svc.name}:${svc.env}`"
        class="service-card"
        :class="{ 'is-offline': svc.health === 'offline', 'is-domain': svc.isDomain }"
      >
        <!-- 头部 -->
        <div class="card-head">
          <div class="card-head-left">
            <span class="svc-icon">{{ svc.icon }}</span>
            <div>
              <div class="svc-name">{{ svc.name }}</div>
              <div class="svc-meta">
                env: {{ svc.env }} · {{ svc.label }}
                <template v-if="svc.endpoint"> · <code>{{ svc.endpoint }}</code></template>
              </div>
            </div>
          </div>
          <span class="health-badge" :class="svc.health">
            <span class="dot"></span>{{ svc.healthLabel }}
          </span>
        </div>

        <!-- 管线: 启动 → 拉取配置 → 本地缓存 -->
        <div class="pipeline">
          <div class="pl-node" :class="{ done: svc.health !== 'offline' }">启动</div>
          <div class="pl-line"></div>
          <div class="pl-node" :class="{ done: svc.health === 'online' }">拉取配置</div>
          <div class="pl-line"></div>
          <div class="pl-node" :class="{ done: svc.health === 'online' }">本地缓存</div>
        </div>

        <!-- 配置绑定展示 -->
        <div class="config-bind" v-if="svc.config && Object.keys(svc.config).length">
          <div class="bind-row" v-for="(val, key) in svc.config" :key="key">
            <span>{{ key }}</span>
            <code>{{ val }}</code>
          </div>
        </div>

        <!-- 空状态 -->
        <div class="status-empty" v-else-if="svc.health !== 'checking'">
          <template v-if="svc.health === 'offline'">
            服务未启动或不可达。请启动对应的 demo-service 实例。
          </template>
          <template v-else>
            该服务已注册配置组，尚未启动对应进程。
          </template>
        </div>
        <div class="status-empty checking" v-else>
          正在检测...
        </div>
      </article>

      <!-- 图例卡 -->
      <article class="service-card muted-legend">
        <div class="svc-name">服务拉取接口</div>
        <p>各领域服务启动时通过共享 ConfigClient SDK 拉取配置：</p>
        <code class="endpoint">GET /api/configs/{service}/{env}/published</code>
        <p style="margin-top:10px">健康检测通过 <code>/health</code> 端点，返回 <code>{"status":"ok","config_from_center":true}</code>。</p>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, reactive } from 'vue'
import { Refresh } from '@element-plus/icons-vue'

const props = defineProps({
  configs: { type: Array, default: () => [] },
})

// ---------------------------------------------------------------------------
// 服务注册表 — 所有可能运行的微服务实例
// ---------------------------------------------------------------------------
const REGISTRY = [
  // 通用服务
  { name: 'order-service',   env: 'dev',  endpoint: 'http://localhost:3000', label: '通用配置回显',      icon: '📦' },
  { name: 'order-service',   env: 'test', endpoint: null,                    label: '通用配置回显',      icon: '📦' },
  { name: 'user-service',    env: 'dev',  endpoint: null,                    label: '通用配置回显',      icon: '📦' },
  { name: 'user-service',    env: 'prod', endpoint: null,                    label: '通用配置回显',      icon: '📦' },
  // 领域仿真服务
  { name: 'radar-service',   env: 'dev',  endpoint: 'http://localhost:3001', label: '火控雷达',          icon: '🔴', isDomain: true },
  { name: 'radar-service',   env: 'prod', endpoint: null,                    label: '火控雷达 (生产)',   icon: '🔴', isDomain: true },
  { name: 'sensor-service',  env: 'dev',  endpoint: 'http://localhost:3002', label: '光电传感器',        icon: '📷', isDomain: true },
  { name: 'sensor-service',  env: 'prod', endpoint: null,                    label: '光电传感器 (生产)', icon: '📷', isDomain: true },
  { name: 'navigation-service', env: 'dev',  endpoint: 'http://localhost:3003', label: '船舶航海',      icon: '🚢', isDomain: true },
  { name: 'navigation-service', env: 'prod', endpoint: null,                    label: '船舶航海 (生产)', icon: '🚢', isDomain: true },
]

const healthMap = reactive({})   // key: "name:env" → 'checking'|'online'|'offline'
const configMap = reactive({})   // key: "name:env" → publishedData or config from /health

// ---------------------------------------------------------------------------
// 服务视图 — 合并注册表 + configs prop + healthMap
// ---------------------------------------------------------------------------
const serviceViews = computed(() => REGISTRY.map(reg => {
  const key = `${reg.name}:${reg.env}`
  const health = healthMap[key] || 'checking'
  const matched = props.configs.find(c => c.service === reg.name && c.env === reg.env)

  let healthLabel = '检测中'
  if (health === 'online')  healthLabel = '在线'
  if (health === 'offline') healthLabel = '离线'
  if (health === 'bound')   healthLabel = '配置已绑定'

  // 配置来源: 优先用 health check 返回的 config，其次用后端 API 的 publishedData
  const config = configMap[key] || matched?.publishedData || null

  return {
    ...reg,
    health,
    healthLabel,
    config,
  }
}))

// ---------------------------------------------------------------------------
// 健康检测
// ---------------------------------------------------------------------------
async function checkHealth(reg) {
  const key = `${reg.name}:${reg.env}`
  if (!reg.endpoint) {
    // 无 endpoint → 仅检查是否有配置绑定
    const matched = props.configs.find(c => c.service === reg.name && c.env === reg.env)
    healthMap[key] = matched ? 'bound' : 'offline'
    return
  }

  healthMap[key] = 'checking'
  try {
    const resp = await fetch(`${reg.endpoint}/health`, { signal: AbortSignal.timeout(3000) })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    const data = await resp.json()

    if (data.config_from_center) {
      healthMap[key] = 'online'
      // 同时拉取该服务的 / 端点获取配置详情
      try {
        const infoResp = await fetch(`${reg.endpoint}/`, { signal: AbortSignal.timeout(2000) })
        if (infoResp.ok) {
          const info = await infoResp.json()
          configMap[key] = info.config || info.raw || {}
        }
      } catch { /* / endpoint optional */ }
    } else {
      healthMap[key] = 'offline'
    }
  } catch {
    healthMap[key] = 'offline'
  }
}

async function checkAll() {
  const tasks = REGISTRY.map(reg => checkHealth(reg))
  await Promise.allSettled(tasks)
}

onMounted(checkAll)
</script>

<style scoped>
.service-status {
  padding: 20px 24px 28px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  flex-shrink: 0;
}
.panel-head h2 { font-size: 18px; font-weight: 650; color: var(--text-primary); margin-bottom: 4px; }
.panel-head p  { font-size: var(--font-size-sm); color: var(--text-secondary); }
.panel-action {
  height: 32px; display: inline-flex; align-items: center; gap: 6px;
  padding: 0 12px; border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md); background: var(--bg-surface);
  color: var(--text-secondary); cursor: pointer; font-family: var(--font-family);
  font-size: var(--font-size-sm); flex-shrink: 0;
}
.panel-action:hover { border-color: var(--border-hover); color: var(--text-primary); }

/* ── 可滚动网格 ── */
.service-scroll {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 14px;
  align-content: start;
  padding-right: 4px;
}

.service-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 18px;
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.service-card.is-domain {
  border-left: 3px solid var(--color-primary);
}

.service-card.is-offline {
  opacity: 0.7;
}

/* ── 卡片头部 ── */
.card-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}
.card-head-left { display: flex; gap: 10px; align-items: flex-start; min-width: 0; }
.svc-icon { font-size: 20px; flex-shrink: 0; margin-top: 1px; }
.svc-name {
  font-size: var(--font-size-base);
  font-weight: 650;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.svc-meta {
  margin-top: 3px;
  font-size: var(--font-size-xs);
  color: var(--text-placeholder);
  line-height: 1.5;
}
.svc-meta code {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 11px;
  background: var(--bg-subtle);
  padding: 1px 5px;
  border-radius: var(--radius-sm);
}

/* ── 健康徽章 ── */
.health-badge {
  display: inline-flex; align-items: center; gap: 6px;
  height: 24px; padding: 0 9px; border-radius: 999px;
  font-size: var(--font-size-xs); font-weight: 600; flex-shrink: 0;
}
.health-badge .dot { width: 7px; height: 7px; border-radius: 50%; }
.health-badge.online  { color: var(--color-published); background: var(--color-published-bg); }
.health-badge.online  .dot { background: var(--color-published); }
.health-badge.offline { color: var(--diff-deleted-text); background: var(--diff-deleted-bg); }
.health-badge.offline .dot { background: var(--diff-deleted-text); }
.health-badge.checking{ color: var(--color-pending); background: var(--color-pending-bg); }
.health-badge.checking .dot { background: var(--color-pending); }
.health-badge.bound   { color: var(--color-primary); background: var(--bg-selected); }
.health-badge.bound   .dot { background: var(--color-primary); }

/* ── 管线 ── */
.pipeline {
  display: flex; align-items: center; gap: 6px;
}
.pl-node {
  height: 26px; display: inline-flex; align-items: center; padding: 0 9px;
  border-radius: var(--radius-md); border: 1px solid var(--border-subtle);
  color: var(--text-placeholder); background: var(--bg-subtle);
  font-size: var(--font-size-xs); font-weight: 600;
}
.pl-node.done {
  color: var(--color-primary); background: var(--bg-selected);
  border-color: rgba(94, 106, 210, .22);
}
.pl-line { height: 1px; flex: 1; background: var(--border-subtle); }

/* ── 配置绑定表 ── */
.config-bind {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}
.bind-row {
  display: flex; justify-content: space-between; gap: 12px;
  padding: 9px 12px; border-bottom: 1px solid var(--border-subtle);
  font-size: var(--font-size-sm);
}
.bind-row:last-child { border-bottom: none; }
.bind-row span { color: var(--text-secondary); flex-shrink: 0; }
.bind-row code {
  color: var(--text-primary); background: var(--bg-subtle);
  border: 1px solid var(--border-subtle); border-radius: var(--radius-sm);
  padding: 2px 6px; font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: var(--font-size-xs); max-width: 60%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* ── 空状态 / 图例卡 ── */
.status-empty {
  color: var(--text-placeholder); font-size: var(--font-size-sm);
  border: 1px dashed var(--border-subtle); border-radius: var(--radius-md);
  padding: 16px; text-align: center; line-height: 1.6;
}
.status-empty.checking { color: var(--color-pending); }

.muted-legend {
  border-style: dashed;
  border-color: var(--border-subtle);
  background: var(--bg-subtle);
}
.muted-legend p {
  color: var(--text-secondary); font-size: var(--font-size-sm); line-height: 1.7; margin: 0;
}
.muted-legend .svc-name {
  margin-bottom: 2px;
}
.endpoint {
  display: inline-block; margin-top: 8px;
  color: var(--text-primary); background: var(--bg-surface);
  border: 1px solid var(--border-subtle); border-radius: var(--radius-sm);
  padding: 4px 8px;
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: var(--font-size-xs);
}
</style>
