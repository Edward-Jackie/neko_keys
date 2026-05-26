package router

import (
	"github.com/gin-gonic/gin"

	"neko_backend/handlers"
	"neko_backend/middleware"
)

// Setup wires all routes onto the given engine.
func Setup(r *gin.Engine, plugin *handlers.PluginHandler, admin *handlers.AdminHandler, jwtSecret string) {
	r.Use(middleware.CORS())

	// Public plugin endpoints — no authentication. Rate-limited per client IP
	// (shared bucket across activate + validate): ~20 req/min, burst 5.
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RateLimit(20, 5))
	{
		v1.POST("/activate", plugin.Activate)
		v1.POST("/validate", plugin.Validate)
	}

	// Admin endpoints — JWT required.
	adminGroup := r.Group("/api/admin")
	{
		adminGroup.POST("/login", admin.Login)

		protected := adminGroup.Group("")
		protected.Use(middleware.AdminAuth(jwtSecret))
		{
			protected.GET("/dashboard", admin.Dashboard)

			protected.GET("/batches", admin.ListBatches)
			protected.POST("/batches", admin.CreateBatch)

			protected.GET("/keys", admin.ListKeys)
			protected.GET("/keys/:id", admin.GetKey)
			protected.PUT("/keys/:id", admin.UpdateKey)
			protected.GET("/keys/:id/logs", admin.GetKeyLogs)

			protected.GET("/alerts", admin.ListAlerts)
		}
	}
}
