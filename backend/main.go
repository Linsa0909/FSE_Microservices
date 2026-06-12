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
		// === 火控雷达 ===
		{
			service: "radar-service",
			env:     "dev",
			data: map[string]string{
				"radar.scan_rate_hz": "60",
				"radar.range_km":     "150",
				"radar.max_targets":  "32",
				"radar.mode":         "search",
				"radar.band":         "X",
			},
		},
		{
			service: "radar-service",
			env:     "prod",
			data: map[string]string{
				"radar.scan_rate_hz": "120",
				"radar.range_km":     "300",
				"radar.max_targets":  "128",
				"radar.mode":         "track",
				"radar.band":         "C",
			},
		},
		// === 光电传感器 ===
		{
			service: "sensor-service",
			env:     "dev",
			data: map[string]string{
				"sensor.wavelength_nm": "1550",
				"sensor.fps":           "30",
				"sensor.resolution":    "1280x720",
				"sensor.sensitivity":   "5",
				"sensor.mode":          "day",
			},
		},
		{
			service: "sensor-service",
			env:     "prod",
			data: map[string]string{
				"sensor.wavelength_nm": "8000",
				"sensor.fps":           "60",
				"sensor.resolution":    "1920x1080",
				"sensor.sensitivity":   "8",
				"sensor.mode":          "thermal",
			},
		},
		// === 船舶航海 ===
		{
			service: "navigation-service",
			env:     "dev",
			data: map[string]string{
				"nav.update_interval_s": "5",
				"nav.speed_knots":       "20",
				"nav.heading_deg":       "0",
				"nav.destination":       "Shanghai",
				"nav.auto_pilot":        "false",
				"nav.waypoints":         "PortA,PortB,PortC",
			},
		},
		{
			service: "navigation-service",
			env:     "prod",
			data: map[string]string{
				"nav.update_interval_s": "2",
				"nav.speed_knots":       "35",
				"nav.heading_deg":       "180",
				"nav.destination":       "Singapore",
				"nav.auto_pilot":        "true",
				"nav.waypoints":         "Shanghai,Xiamen,Singapore",
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
