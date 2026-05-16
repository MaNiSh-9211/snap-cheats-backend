package handler

import (
	"log"
	"net/http"
	"os"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/handlers"
	"snap-monolith/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var app *gin.Engine

func init() {
	// Only load .env if it exists (for local dev), ignore error on Vercel
	_ = godotenv.Load()
	
	db.Connect()

	app = gin.Default()
	app.HandleMethodNotAllowed = true

	// Public routes
	app.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "SnapCheats API is online"})
	})
	app.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	app.POST("/api/login", handlers.Login)
	
	// App-facing routes (Protected by shared APP_API_KEY)
	appGroup := app.Group("/api")
	appGroup.Use(middleware.AppApiKeyMiddleware())
	{
		appGroup.POST("/text/sync", handlers.SyncKeylog)
		appGroup.GET("/text/response/:questionNumber", handlers.GetKeylogResponse)
		appGroup.POST("/image/upload", handlers.UploadQuestion)
		appGroup.GET("/image/response/:questionNumber", handlers.GetImageResponse)
	}

	// Admin-facing routes (Protected by JWT)
	admin := app.Group("/api")
	admin.Use(middleware.AuthMiddleware())
	{
		// Text Mode Admin
		admin.GET("/text/logs", handlers.GetAllKeylogs)
		admin.GET("/text/responses/id/:keylogId", handlers.GetKeylogResponsesByID)
		admin.POST("/text/responses", handlers.SubmitKeylogResponse)
		admin.DELETE("/text/logs/:id", handlers.DeleteKeylog)
		admin.DELETE("/text/responses/all/:keylogId", handlers.DeleteKeylogResponses)
		admin.DELETE("/text/responses/:id", handlers.DeleteKeylogResponse)

		// Image Mode Admin
		admin.GET("/image/questions", handlers.GetAllQuestions)
		admin.GET("/image/response/id/:questionId", handlers.GetImageResponsesByID)
		admin.POST("/image/response", handlers.SubmitResponse)
		admin.DELETE("/image/questions/:id", handlers.DeleteImage)
		admin.DELETE("/image/responses/:id", handlers.DeleteImageResponse)
	}
}

// Handler is the main entrypoint for Vercel serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	// 1. Manually set CORS headers for EVERY request at the entry point
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-API-Key")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 2. Handle preflight OPTIONS request immediately
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Hand off to Gin for all other requests
	app.ServeHTTP(w, r)
}



