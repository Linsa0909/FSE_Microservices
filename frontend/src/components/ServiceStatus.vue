<template>
  <section class="service-status">
    <div class="panel-head">
      <div>
        <h2>微服务状态</h2>
        <p>探测各领域微服务 /health 端点，验证配置拉取与运行状态。</p>
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
        :class="{ 'is-offline': svc.health === 'offline' }"
      >
        <!-- 头部 -->
        <div class="card-head">
          <div class="card-head-left">
            <span class="svc-icon">{{ svc.icon }}</span>
            <div>
              <div class="svc-name">{{ svc.label }}</div>
              <div class="svc-meta">
                {{ svc.name }} · env: {{ svc.env }} · <code>{{ svc.endpoint }}</code>
              </div>
            </div>
          </div>
          <span class="health-badge" :class="svc.health">
            <span class="dot"></span>{{ svc.healthLabel }}
          </span>
        </div>

        <!-- 管线: 仅运行中/离线服务展示 -->
        <div class="pipeline">
          <div class="pl-node" :class="{ done: svc.health !== 'offline' }">1. 启动进程</div>
          <div class="pl-line"></div>
          <div class="pl-node" :class="{ done: svc.health === 'online' }">2. 拉取配置</div>
          <div class="pl-line"></div>
          <div class="pl-node" :class="{ done: svc.health === 'online' }">3. 本地缓存</div>
        </div>

        <!-- 配置详情: 在线 → 展示拉取到的配置; 离线 → 展示后端种子配置 -->
        <div class="config-bind" v-if="svc.displayConfig && Object.keys(svc.displayConfig).length">
          <div class="bind-row" v-for="(val, key) in svc.displayConfig" :key="key">
            <span>{{ key }}</span>
            <code>{{ val }}</code>
          </div>
        </div>

        <div class="status-empty" v-else-if="svc.health === 'offline'">
          服务未启动或不可达。<br/>请执行：<code>SERVICE_TYPE={{ svc.type }} CONFIG_SERVICE={{ svc.name }} {{ svc.portEnv }}={{ svc.port }} go run .</code>
        </div>
        <div class="status-empty checking" v-else>
          正在检测...
        </div>
      </article>

      <!-- 图例卡 -->
      <article class="service-card muted-legend">
        <div class="svc-name">健康检测说明</div>
        <p>前端直接探测每个服务的 <code>/health</code> 端点（CORS 已开启）。</p>
        <code class="endpoint">GET http://localhost:3001/health → {"status":"ok","config_from_center":true}</code>
        <p style="margin-top:10px">
          <code>config_from_center=true</code> 表示服务成功从配置中心拉取了 PublishedData。<br/>
          所有服务均支持<strong>降级启动</strong>：配置中心不可达时使用默认配置，<code>/health</code> 返回 <code>degraded</code>。
        </p>
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
// 服务注册表 — 仅注册有实际 endpoint 的可探测服务
// ---------------------------------------------------------------------------
const REGISTRY = [
  { name: 'order-service',   env: 'dev', endpoint: 'http://localhost:3000', label: '通用配置回显',      icon: '📦', type: 'default',    port: '3000', portEnv: 'PORT' },
  { name: 'radar-service',   env: 'dev', endpoint: 'http://localhost:3001', label: '火控雷达',          icon: '🔴', type: 'radar',      port: '3001', portEnv: 'RADAR_PORT' },
  { name: 'sensor-service',  env: 'dev', endpoint: 'http://localhost:3002', label: '光电传感器',        icon: '📷', type: 'sensor',     port: '3002', portEnv: 'SENSOR_PORT' },
  { name: 'navigation-service', env: 'dev', endpoint: 'http://localhost:3003', label: '船舶航海',     icon: '🚢', type: 'navigation', port: '3003', portEnv: 'NAV_PORT' },
]

const healthMap = reactive({})
const liveConfigMap = reactive({})  // 在线服务拉取到的实际配置

// ---------------------------------------------------------------------------
// 服务视图
// ---------------------------------------------------------------------------
const serviceViews = computed(() => REGISTRY.map(reg => {
  const key = `${reg.name}:${reg.env}`
  const health = healthMap[key] || 'checking'
  const matched = props.configs.find(c => c.service === reg.name && c.env === reg.env)

  let healthLabel = '检测中'
  if (health === 'online')  healthLabel = '在线'
  if (health === 'offline') healthLabel = '离线'

  // 配置展示: 在线 → 实际拉取到的配置; 离线 → 后端种子配置 (仅供参考)
  const displayConfig = health === 'online'
    ? liveConfigMap[key]
    : (matched?.publishedData || null)

  return { ...reg, health, healthLabel, displayConfig }
}))

// ---------------------------------------------------------------------------
// 健康检测
// ---------------------------------------------------------------------------
async function checkHealth(reg) {
  const key = `${reg.name}:${reg.env}`

  healthMap[key] = 'checking'
  try {
    const resp = await fetch(`${reg.endpoint}/health`, { signal: AbortSignal.timeout(3000) })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    const data = await resp.json()

    if (data.config_from_center) {
      healthMap[key] = 'online'
      // 拉取该服务实际生效的配置
      try {
        const infoResp = await fetch(`${reg.endpoint}/`, { signal: AbortSignal.timeout(2000) })
        if (infoResp.ok) {
          const info = await infoResp.json()
          liveConfigMap[key] = info.config || info.raw || {}
        }
      } catch { /* GET / 可选 */ }
    } else {
      // config_from_center=false → 降级启动
      healthMap[key] = 'online'
      liveConfigMap[key] = {}
    }
  } catch {
    healthMap[key] = 'offline'
    liveConfigMap[key] = null
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
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
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
  border-left: 3px solid var(--color-primary);
}

.service-card.is-offline {
  opacity: 0.65;
  border-left-color: var(--border-subtle);
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

/* ── 管线 ── */
.pipeline {
  display: flex; align-items: center; gap: 6px;
}
.pl-node {
  height: 26px; display: inline-flex; align-items: center; padding: 0 10px;
  border-radius: var(--radius-md); border: 1px solid var(--border-subtle);
  color: var(--text-placeholder); background: var(--bg-subtle);
  font-size: var(--font-size-xs); font-weight: 600; white-space: nowrap;
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

/* ── 空状态 / 图例 ── */
.status-empty {
  color: var(--text-placeholder); font-size: var(--font-size-sm);
  border: 1px dashed var(--border-subtle); border-radius: var(--radius-md);
  padding: 16px; text-align: center; line-height: 1.8;
}
.status-empty.checking { color: var(--color-pending); }
.status-empty code {
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px;
  background: var(--bg-subtle); padding: 1px 5px; border-radius: var(--radius-sm);
}

.muted-legend {
  border-style: dashed;
  border-left: 3px dashed var(--border-subtle);
  border-color: var(--border-subtle);
  background: var(--bg-subtle);
}
.muted-legend p {
  color: var(--text-secondary); font-size: var(--font-size-sm); line-height: 1.7; margin: 0;
}
.muted-legend .svc-name { margin-bottom: 2px; }
.endpoint {
  display: inline-block; margin-top: 8px;
  color: var(--text-primary); background: var(--bg-surface);
  border: 1px solid var(--border-subtle); border-radius: var(--radius-sm);
  padding: 4px 8px;
  font-family: 'SF Mono', 'Fira Code', monospace; font-size: var(--font-size-xs);
}
</style>
