package api

import (

	"github.com/gin-gonic/gin"
	"skynet-net-engine-api/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "skynet-net-engine-api/docs" // Import generated docs
)

func Start(port string) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	
	// Middleware
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		// CORS for Dev
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-App-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		logger.Info("API Request", 
			logger.Field("method", c.Request.Method),
			logger.Field("path", c.Request.URL.Path),
			logger.Field("ip", c.ClientIP()),
		)
		c.Next()
	})


	// API Routes (must come before NoRoute)
	// ... (V1 routes defined below) ...

	// Global 404 Handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Route Not Found"})
	})

	// Public V1 Routes
	v1 := r.Group("/api/v1")
	{
		// Documentation (Public)
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		v1.GET("/health", HealthCheck) // Health check is often public too
	}

	// Secured V1 Routes
	secured := v1.Group("/")
	secured.Use(func(c *gin.Context) {
		key := c.GetHeader("X-App-Key")
		// Hardcoded for MVP, move to ENV
		if key != "netengine_secret_key_123" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	})
	{
		// Internal Control
		secured.POST("/sync/:id", SyncRouter)
		secured.POST("/kick", KickUser)

		// CRUD Bridge
		secured.POST("/secret", CreateSecret)
		secured.PUT("/secret/:user", UpdatePlan)
		
		// Advanced
		secured.POST("/isolate", IsolateUser)
		secured.GET("/routers", GetRouters)
		secured.GET("/monitoring/targets", GetTargets)
		secured.GET("/router/:id/health", GetRouterHealth)
		secured.GET("/router/:id/users", GetAllUsers)
		secured.GET("/router/:id/traffic", GetUserTraffic)
		secured.POST("/router/:id/backup", TriggerBackup)
	}

	logger.Info("Starting API Server on " + port)
	if err := r.Run(port); err != nil {
		logger.Fatal("Failed to start API server")
	}
}
