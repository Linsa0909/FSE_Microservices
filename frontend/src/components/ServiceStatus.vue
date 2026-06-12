<template>
  <section class="service-status">
    <div class="panel-head">
      <div>
        <h2>微服务状态</h2>
        <p>验证示例微服务启动后是否成功拉取配置中心的 PublishedData。</p>
      </div>
      <button class="panel-action" @click="checkDemoService">
        <el-icon><Refresh /></el-icon>
        刷新状态
      </button>
    </div>

    <div class="service-grid">
      <article v-for="service in serviceViews" :key="`${service.name}:${service.env}`" class="service-card">
        <div class="service-top">
          <div>
            <div class="service-name">{{ service.name }}</div>
            <div class="service-meta">env: {{ service.env }} · endpoint: {{ service.endpoint }}</div>
          </div>
          <span class="health" :class="service.status">
            <span></span>{{ service.statusLabel }}
          </span>
        </div>

        <div class="service-pipeline">
          <div class="pipe-node done">启动</div>
          <div class="pipe-line"></div>
          <div class="pipe-node" :class="{ done: service.status === 'online' }">拉取配置</div>
          <div class="pipe-line"></div>
          <div class="pipe-node" :class="{ done: service.status === 'online' }">本地缓存</div>
        </div>

        <div class="config-cache" v-if="service.config">
          <div class="cache-row">
            <span>数据库连接</span>
            <code>{{ service.config['db.url'] || '未返回' }}</code>
          </div>
          <div class="cache-row">
            <span>服务端口</span>
            <code>{{ service.config['server.port'] || service.defaultPort }}</code>
          </div>
          <div class="cache-row">
            <span>日志级别</span>
            <code>{{ service.config['log.level'] || '未返回' }}</code>
          </div>
        </div>

        <div class="status-empty" v-else>
          {{ service.emptyText }}
        </div>
      </article>

      <article class="service-card muted">
        <div class="service-name">客户端拉取接口</div>
        <p>示例服务启动时调用配置中心：</p>
        <code class="endpoint">GET /api/configs/order-service/dev/published</code>
        <p>当前 MVP 使用启动拉取 + 前端轮询，运行期动态推送通过预留接口扩展。</p>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'

const demoStatus = ref('checking')
const demoConfig = ref(null)

const serviceSeeds = [
  { name: 'order-service', env: 'dev', endpoint: 'http://localhost:3000', defaultPort: '3000', realDemo: true },
  { name: 'order-service', env: 'test', endpoint: '待接入', defaultPort: '3000', realDemo: false },
  { name: 'user-service', env: 'dev', endpoint: '待接入', defaultPort: '3001', realDemo: false },
]

const statusLabel = computed(() => {
  if (demoStatus.value === 'online') return '在线'
  if (demoStatus.value === 'checking') return '检测中'
  return '离线'
})

const props = defineProps({
  configs: { type: Array, default: () => [] },
})

const serviceViews = computed(() => serviceSeeds.map(seed => {
  if (seed.realDemo) {
    return {
      ...seed,
      status: demoStatus.value,
      statusLabel: statusLabel.value,
      config: demoConfig.value,
      emptyText: demoStatus.value === 'checking' ? '正在检测示例服务...' : '未获取到示例服务配置。请先启动后端和 demo-service。',
    }
  }

  const matched = props.configs.find(item => item.service === seed.name && item.env === seed.env)
  return {
    ...seed,
    status: matched ? 'reserved' : 'offline',
    statusLabel: matched ? '配置已绑定' : '未配置',
    config: matched?.publishedData || null,
    emptyText: matched ? '当前仅展示配置绑定，尚未启动对应示例服务进程。' : '未找到对应配置组。',
  }
}))

async function checkDemoService() {
  demoStatus.value = 'checking'
  try {
    const resp = await fetch('/demo/')
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    const data = await resp.json()
    demoConfig.value = data.config || {}
    demoStatus.value = 'online'
  } catch {
    demoConfig.value = null
    demoStatus.value = 'offline'
  }
}

onMounted(checkDemoService)
</script>

<style scoped>
.service-status { padding: 20px 24px 28px; overflow: auto; }
.panel-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.panel-head h2 { font-size: 18px; font-weight: 650; color: var(--text-primary); margin-bottom: 4px; }
.panel-head p { font-size: var(--font-size-sm); color: var(--text-secondary); }
.panel-action { height: 32px; display: inline-flex; align-items: center; gap: 6px; padding: 0 12px; border: 1px solid var(--border-subtle); border-radius: var(--radius-md); background: var(--bg-surface); color: var(--text-secondary); cursor: pointer; font-family: var(--font-family); }
.panel-action:hover { border-color: var(--border-hover); color: var(--text-primary); }
.service-grid { display: grid; grid-template-columns: minmax(0, 1.25fr) minmax(320px, .75fr); gap: 16px; }
.service-card { background: var(--bg-surface); border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); padding: 16px; box-shadow: var(--shadow-sm); }
.service-card.muted p { color: var(--text-secondary); font-size: var(--font-size-sm); line-height: 1.7; margin-top: 10px; }
.service-top { display: flex; justify-content: space-between; gap: 16px; align-items: flex-start; margin-bottom: 18px; }
.service-name { font-size: var(--font-size-lg); font-weight: 650; color: var(--text-primary); }
.service-meta { margin-top: 4px; font-size: var(--font-size-xs); color: var(--text-placeholder); }
.health { display: inline-flex; align-items: center; gap: 6px; height: 24px; padding: 0 9px; border-radius: 999px; font-size: var(--font-size-xs); font-weight: 600; }
.health span { width: 7px; height: 7px; border-radius: 50%; }
.health.online { color: var(--color-published); background: var(--color-published-bg); }
.health.online span { background: var(--color-published); }
.health.offline { color: var(--diff-deleted-text); background: var(--diff-deleted-bg); }
.health.offline span { background: var(--diff-deleted-text); }
.health.checking { color: var(--color-pending); background: var(--color-pending-bg); }
.health.checking span { background: var(--color-pending); }
.health.reserved { color: var(--color-primary); background: var(--bg-selected); }
.health.reserved span { background: var(--color-primary); }
.service-pipeline { display: flex; align-items: center; gap: 8px; margin-bottom: 18px; }
.pipe-node { height: 28px; display: inline-flex; align-items: center; padding: 0 10px; border-radius: var(--radius-md); border: 1px solid var(--border-subtle); color: var(--text-placeholder); background: var(--bg-subtle); font-size: var(--font-size-xs); font-weight: 600; }
.pipe-node.done { color: var(--color-primary); background: var(--bg-selected); border-color: rgba(94, 106, 210, .22); }
.pipe-line { height: 1px; flex: 1; background: var(--border-subtle); }
.config-cache { border: 1px solid var(--border-subtle); border-radius: var(--radius-md); overflow: hidden; }
.cache-row { display: flex; justify-content: space-between; gap: 16px; padding: 10px 12px; border-bottom: 1px solid var(--border-subtle); font-size: var(--font-size-sm); }
.cache-row:last-child { border-bottom: none; }
.cache-row span { color: var(--text-secondary); }
.cache-row code, .endpoint { color: var(--text-primary); background: var(--bg-subtle); border: 1px solid var(--border-subtle); border-radius: var(--radius-sm); padding: 2px 6px; font-family: 'SF Mono', 'Fira Code', monospace; font-size: var(--font-size-xs); }
.endpoint { display: inline-block; margin-top: 8px; }
.status-empty { color: var(--text-placeholder); font-size: var(--font-size-sm); border: 1px dashed var(--border-subtle); border-radius: var(--radius-md); padding: 18px; text-align: center; }
@media (max-width: 920px) {
  .service-grid { grid-template-columns: 1fr; }
}
</style>
