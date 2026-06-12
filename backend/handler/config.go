package handler

import (
	"net/http"
	"strings"

	"config-center/store"

	"github.com/gin-gonic/gin"
)

// Handler 持有 ConfigStore 引用
type Handler struct {
	Store *store.ConfigStore
}

// New 创建 Handler
func New(s *store.ConfigStore) *Handler {
	return &Handler{Store: s}
}

// --- 请求体 ---

type SetKeyBody struct {
	Value string `json:"value" binding:"required"`
}

// --- 1. GET /api/configs — 所有服务配置列表 ---
func (h *Handler) ListAll(c *gin.Context) {
	h.Store.RLock()
	defer h.Store.RUnlock()

	result := make([]*store.ConfigGroup, 0, len(h.Store.Configs))
	for _, g := range h.Store.Configs {
		result = append(result, g)
	}
	c.JSON(http.StatusOK, gin.H{"configs": result})
}

// --- 2. GET /api/configs/:service/:env — 单个 ConfigGroup + ChangeLog ---
func (h *Handler) GetOne(c *gin.Context) {
	h.Store.RLock()
	defer h.Store.RUnlock()

	service := c.Param("service")
	env := c.Param("env")
	g := h.Store.Get(service, env)
	if g == nil {
		// 不存在则返回空配置组
		g = &store.ConfigGroup{
			Service:       service,
			Env:           env,
			PublishedData: make(map[string]string),
			DraftData:     make(map[string]string),
		}
	}

	logs := h.Store.GetLogs(service, env)
	if logs == nil {
		logs = []store.ChangeRecord{}
	}

	c.JSON(http.StatusOK, gin.H{
		"config": g,
		"logs":   logs,
	})
}

// --- 3. PUT /api/configs/:service/:env/keys/:key — 新增/修改配置项 ---
func (h *Handler) SetKey(c *gin.Context) {
	var body SetKeyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value is required"})
		return
	}

	h.Store.Lock()
	defer h.Store.Unlock()

	service := c.Param("service")
	env := c.Param("env")
	key := c.Param("key")

	h.Store.SetKey(service, env, key, body.Value)
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// --- 4. DELETE /api/configs/:service/:env/keys/:key — 删除配置项 ---
func (h *Handler) DeleteKey(c *gin.Context) {
	h.Store.Lock()
	defer h.Store.Unlock()

	service := c.Param("service")
	env := c.Param("env")
	key := c.Param("key")

	h.Store.DeleteKey(service, env, key)
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// --- 5. POST /api/configs/:service/:env/publish — 发布配置 ---
func (h *Handler) Publish(c *gin.Context) {
	h.Store.Lock()
	defer h.Store.Unlock()

	service := c.Param("service")
	env := c.Param("env")

	h.Store.Publish(service, env)
	c.JSON(http.StatusOK, gin.H{"message": "published"})
}

// --- 6. GET /api/configs/:service/:env/published — 仅返回 PublishedData（微服务用） ---
func (h *Handler) GetPublished(c *gin.Context) {
	h.Store.RLock()
	defer h.Store.RUnlock()

	service := c.Param("service")
	env := c.Param("env")
	g := h.Store.Get(service, env)
	if g == nil {
		c.JSON(http.StatusOK, gin.H{"config": make(map[string]string)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": g.PublishedData})
}

// --- 7. POST /api/configs/:service/:env/push — 配置推送预留接口 ---
func (h *Handler) Push(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "push endpoint reserved for future dynamic push"})
}

// --- 8. GET /api/configs/watch — SSE 推送预留端点 ---
func (h *Handler) Watch(c *gin.Context) {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/event-stream") {
		// SSE 客户端：返回提示后立即关闭
		c.String(http.StatusOK, "data: {\"message\":\"SSE watch endpoint reserved for future dynamic push\"}\n\n")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SSE watch endpoint reserved for future dynamic push"})
}
