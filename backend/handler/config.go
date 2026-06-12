// Package handler implements HTTP handlers for the config center REST API.
// It provides 8 endpoints: list, get, set, delete, publish, get-published, push (reserved), and watch (reserved).
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

type setKeyBody struct {
	Value string `json:"value"`
}

// ListAll handles GET /api/configs — returns all config groups.
func (h *Handler) ListAll(c *gin.Context) {
	h.Store.RLock()
	defer h.Store.RUnlock()

	result := make([]*store.ConfigGroup, 0, len(h.Store.Configs))
	for _, g := range h.Store.Configs {
		result = append(result, g)
	}
	c.JSON(http.StatusOK, gin.H{"configs": result})
}

// GetOne handles GET /api/configs/:service/:env — returns a single ConfigGroup with ChangeLog.
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

// SetKey handles PUT /api/configs/:service/:env/keys/:key — adds or updates a key in DraftData.
func (h *Handler) SetKey(c *gin.Context) {
	var body setKeyBody
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

// DeleteKey handles DELETE /api/configs/:service/:env/keys/:key — removes a key from DraftData.
func (h *Handler) DeleteKey(c *gin.Context) {
	h.Store.Lock()
	defer h.Store.Unlock()

	service := c.Param("service")
	env := c.Param("env")
	key := c.Param("key")

	h.Store.DeleteKey(service, env, key)
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// Publish handles POST /api/configs/:service/:env/publish — deep-copies DraftData to PublishedData.
func (h *Handler) Publish(c *gin.Context) {
	h.Store.Lock()
	defer h.Store.Unlock()

	service := c.Param("service")
	env := c.Param("env")

	h.Store.Publish(service, env)
	c.JSON(http.StatusOK, gin.H{"message": "published"})
}

// GetPublished handles GET /api/configs/:service/:env/published — returns only PublishedData for microservice consumption.
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

// Push handles POST /api/configs/:service/:env/push — reserved endpoint for future dynamic config push.
func (h *Handler) Push(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "push endpoint reserved for future dynamic push"})
}

// Watch handles GET /api/configs/watch — reserved SSE endpoint for future dynamic config push.
func (h *Handler) Watch(c *gin.Context) {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/event-stream") {
		// SSE 客户端：返回提示后立即关闭
		c.String(http.StatusOK, "data: {\"message\":\"SSE watch endpoint reserved for future dynamic push\"}\n\n")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SSE watch endpoint reserved for future dynamic push"})
}
