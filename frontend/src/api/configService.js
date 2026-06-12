import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 5000,
})

// GET /api/configs — 所有配置列表
export async function getAllConfigs() {
  const { data } = await api.get('/configs')
  return data.configs || []
}

// GET /api/configs/:service/:env — 单个配置组 + 变更日志
export async function getOneConfig(service, env) {
  const { data } = await api.get(`/configs/${service}/${env}`)
  return data
}

// PUT /api/configs/:service/:env/keys/:key — 新增/修改配置项
export async function setKey(service, env, key, value) {
  await api.put(`/configs/${service}/${env}/keys/${encodeURIComponent(key)}`, { value })
}

// DELETE /api/configs/:service/:env/keys/:key — 删除配置项
export async function deleteKey(service, env, key) {
  await api.delete(`/configs/${service}/${env}/keys/${encodeURIComponent(key)}`)
}

// POST /api/configs/:service/:env/publish — 发布配置
export async function publishConfig(service, env) {
  await api.post(`/configs/${service}/${env}/publish`)
}
