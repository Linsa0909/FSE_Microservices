// Package main is the entry point for the config center backend.
// It initializes the in-memory config store, sets up seed data, registers all HTTP routes, and starts the Gin server on :8080.
package main

import (
	"log"

	"config-center/handler"
	"config-center/store"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func initSeedData(cs *store.ConfigStore) {
	type seed struct {
		service string
		env     string
		data    map[string]string
	}

	seeds := []seed{
		{
			service: "order-service",
			env:     "dev",
			data: map[string]string{
				"db.url":      "localhost:3306",
				"server.port": "3000",
				"log.level":   "debug",
			},
		},
		{
			service: "order-service",
			env:     "test",
			data: map[string]string{
				"db.url":      "10.0.0.1:3306",
				"server.port": "3000",
				"log.level":   "info",
			},
		},
		{
			service: "user-service",
			env:     "dev",
			data: map[string]string{
				"db.url":      "localhost:3306",
				"server.port": "3001",
				"log.level":   "debug",
			},
		},
		{
			service: "user-service",
			env:     "prod",
			data: map[string]string{
				"db.url":      "prod-db.internal:3306",
				"server.port": "3001",
				"log.level":   "info",
			},
		},
	}

	for _, s := range seeds {
		g := cs.GetOrCreate(s.service, s.env)
		g.DraftData = store.CloneMap(s.data)
		cs.Publish(s.service, s.env) // 初始状态：Draft == Published
	}

	log.Printf("Seed data initialized: %d config groups", len(seeds))
}

func main() {
	// 初始化存储
	cs := store.New()
	initSeedData(cs)

	// 初始化 handler
	h := handler.New(cs)

	// Gin 引擎
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	// 注册路由
	api := r.Group("/api")
	{
		configs := api.Group("/configs")
		{
			configs.GET("", h.ListAll)
			configs.GET("/watch", h.Watch)
			configs.GET("/:service/:env", h.GetOne)
			configs.GET("/:service/:env/published", h.GetPublished)
			configs.PUT("/:service/:env/keys/:key", h.SetKey)
			configs.DELETE("/:service/:env/keys/:key", h.DeleteKey)
			configs.POST("/:service/:env/publish", h.Publish)
			configs.POST("/:service/:env/push", h.Push)
		}
	}

	log.Println("Config Center backend starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
