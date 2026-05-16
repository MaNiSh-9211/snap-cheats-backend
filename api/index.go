package handler

import (
	"net/http"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/handlers"
	"snap-monolith/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var app *gin.Engine

func init() {
	_ = godotenv.Load()
	db.Connect()

	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	// 1. UNIVERSAL CORS (Must be first)
	app.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 2. PUBLIC ROUTES
	app.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"status": "SnapCheats API Online"}) })
	app.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	app.POST("/api/login", handlers.Login)
	app.POST("/login", handlers.Login)

	// 3. APP ROUTES
	appGroup := app.Group("/api")
	appGroup.Use(middleware.AppApiKeyMiddleware())
	{
		appGroup.POST("/text/sync", handlers.SyncKeylog)
		appGroup.GET("/text/response/:questionNumber", handlers.GetKeylogResponse)
		appGroup.POST("/image/upload", handlers.UploadQuestion)
		appGroup.GET("/image/response/:questionNumber", handlers.GetImageResponse)
	}

	// 4. ADMIN ROUTES
	adminGroup := app.Group("/api")
	adminGroup.Use(middleware.AuthMiddleware())
	{
		// Text Admin
		adminGroup.GET("/text/logs", handlers.GetAllKeylogs)
		adminGroup.GET("/text/responses/id/:keylogId", handlers.GetKeylogResponsesByID)
		adminGroup.POST("/text/responses", handlers.SubmitKeylogResponse)
		adminGroup.DELETE("/text/logs/:id", handlers.DeleteKeylog)
		adminGroup.DELETE("/text/responses/all/:keylogId", handlers.DeleteKeylogResponses)
		adminGroup.DELETE("/text/responses/:id", handlers.DeleteKeylogResponse)

		// Image Admin
		adminGroup.GET("/image/questions", handlers.GetAllQuestions)
		adminGroup.GET("/image/response/id/:questionId", handlers.GetImageResponsesByID)
		adminGroup.POST("/image/response", handlers.SubmitResponse)
		adminGroup.DELETE("/image/questions/:id", handlers.DeleteImage)
		adminGroup.DELETE("/image/responses/:id", handlers.DeleteImageResponse)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
