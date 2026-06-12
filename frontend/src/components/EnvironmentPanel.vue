<template>
  <section class="environment-panel">
    <div class="panel-head">
      <div>
        <h2>环境与推送</h2>
        <p>按环境隔离配置组，并展示后续动态推送能力的预留接口。</p>
      </div>
    </div>

    <div class="env-grid">
      <article v-for="env in envCards" :key="env.key" class="env-card">
        <div class="env-card-top">
          <span class="env-name">{{ env.key }}</span>
          <span class="env-count">{{ env.count }} 个配置组</span>
        </div>
        <div class="env-services">
          <span v-for="svc in env.services" :key="svc" class="service-chip">{{ svc }}</span>
          <span v-if="env.services.length === 0" class="empty-chip">暂无服务</span>
        </div>
      </article>
    </div>

    <div class="extension-grid">
      <article class="extension-card">
        <div class="extension-title">配置推送预留</div>
        <p>当前 MVP 通过启动拉取和前端 5 秒轮询保证展示最新配置。后续可将预留接口扩展为 SSE、WebSocket 或长轮询。</p>
        <code>POST /api/configs/:service/:env/push</code>
      </article>
      <article class="extension-card">
        <div class="extension-title">Watch 预留端点</div>
        <p>用于后续客户端订阅配置变更事件。当前返回 reserved 提示，不承载真实长连接逻辑。</p>
        <code>GET /api/configs/watch</code>
      </article>
      <article class="extension-card">
        <div class="extension-title">隔离模型</div>
        <p>当前以 service + env 作为隔离键，例如 order-service:dev。后续可扩展 Namespace / Group / DataId。</p>
        <code>ConfigStore key = service:env</code>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  configs: { type: Array, default: () => [] },
})

const envCards = computed(() => {
  const names = new Set(['dev', 'test', 'prod'])
  props.configs.forEach(c => names.add(c.env))

  return [...names].map(key => {
    const rows = props.configs.filter(c => c.env === key)
    return {
      key,
      count: rows.length,
      services: rows.map(c => c.service),
    }
  })
})
</script>

<style scoped>
.environment-panel { padding: 20px 24px 28px; overflow: auto; }
.panel-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.panel-head h2 { font-size: 18px; font-weight: 650; color: var(--text-primary); margin-bottom: 4px; }
.panel-head p { font-size: var(--font-size-sm); color: var(--text-secondary); }
.env-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; margin-bottom: 16px; }
.env-card, .extension-card { background: var(--bg-surface); border: 1px solid var(--border-subtle); border-radius: var(--radius-lg); padding: 15px; box-shadow: var(--shadow-sm); }
.env-card-top { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 12px; }
.env-name { font-size: 16px; font-weight: 650; color: var(--text-primary); }
.env-count { font-size: var(--font-size-xs); color: var(--text-placeholder); background: var(--bg-subtle); border-radius: 999px; padding: 2px 8px; }
.env-services { display: flex; flex-wrap: wrap; gap: 6px; }
.service-chip, .empty-chip { font-size: var(--font-size-xs); border-radius: var(--radius-sm); padding: 3px 7px; }
.service-chip { color: var(--color-primary); background: var(--bg-selected); }
.empty-chip { color: var(--text-placeholder); background: var(--bg-subtle); }
.extension-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; }
.extension-title { font-size: var(--font-size-base); font-weight: 650; color: var(--text-primary); margin-bottom: 8px; }
.extension-card p { color: var(--text-secondary); font-size: var(--font-size-sm); line-height: 1.7; margin-bottom: 10px; }
.extension-card code { display: inline-block; max-width: 100%; color: var(--text-primary); background: var(--bg-subtle); border: 1px solid var(--border-subtle); border-radius: var(--radius-sm); padding: 3px 7px; font-family: 'SF Mono', 'Fira Code', monospace; font-size: var(--font-size-xs); overflow: hidden; text-overflow: ellipsis; }
@media (max-width: 960px) {
  .env-grid, .extension-grid { grid-template-columns: 1fr; }
}
</style>
