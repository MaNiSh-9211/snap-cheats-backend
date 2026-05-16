package handler

import (
	"net/http"
	"os"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/handlers"
	"snap-monolith/backend/internal/middleware"


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

	// UNIVERSAL CORS: Echoes the requesting Origin back to the browser.
	// This is the "Allow All Origins" strategy that works with credentials.
	app.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Requested-With")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public routes
	app.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "SnapCheats API is online", "version": "1.1.0"})
	})
	
	// Health check (handle both paths)
	health := func(c *gin.Context) { c.JSON(200, gin.H{"status": "healthy"}) }
	app.GET("/health", health)
	app.GET("/api/health", health)
	
	// Login (handle both paths)
	app.POST("/login", handlers.Login)
	app.POST("/api/login", handlers.Login)
	
	// App-facing routes
	appRoutes := []string{"/api", ""}
	for _, prefix := range appRoutes {
		group := app.Group(prefix)
		group.Use(middleware.AppApiKeyMiddleware())
		{
			group.POST("/text/sync", handlers.SyncKeylog)
			group.GET("/text/response/:questionNumber", handlers.GetKeylogResponse)
			group.POST("/image/upload", handlers.UploadQuestion)
			group.GET("/image/response/:questionNumber", handlers.GetImageResponse)
		}
	}

	// Admin-facing routes
	adminRoutes := []string{"/api", ""}
	for _, prefix := range adminRoutes {
		group := app.Group(prefix)
		group.Use(middleware.AuthMiddleware())
		{
			// Text Mode Admin
			group.GET("/text/logs", handlers.GetAllKeylogs)
			group.GET("/text/responses/id/:keylogId", handlers.GetKeylogResponsesByID)
			group.POST("/text/responses", handlers.SubmitKeylogResponse)
			group.DELETE("/text/logs/:id", handlers.DeleteKeylog)
			group.DELETE("/text/responses/all/:keylogId", handlers.DeleteKeylogResponses)
			group.DELETE("/text/responses/:id", handlers.DeleteKeylogResponse)

			// Image Mode Admin
			group.GET("/image/questions", handlers.GetAllQuestions)
			group.GET("/image/response/id/:questionId", handlers.GetImageResponsesByID)
			group.POST("/image/response", handlers.SubmitResponse)
			group.DELETE("/image/questions/:id", handlers.DeleteImage)
			group.DELETE("/image/responses/:id", handlers.DeleteImageResponse)
		}
	}
}

// Handler is the main entrypoint for Vercel serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
