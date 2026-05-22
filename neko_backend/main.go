package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"neko_backend/cache"
	"neko_backend/config"
	"neko_backend/database"
	"neko_backend/handlers"
	"neko_backend/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := database.Init(cfg); err != nil {
		log.Fatalf("init database: %v", err)
	}

	kc := cache.NewKeyCache()

	pluginHandler := handlers.NewPluginHandler(kc)
	adminHandler := handlers.NewAdminHandler(cfg.JWTSecret, kc)

	r := gin.Default()
	router.Setup(r, pluginHandler, adminHandler, cfg.JWTSecret)

	log.Printf("neko_backend listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
