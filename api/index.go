package handler

import (
	"net/http"
	"os"
	"strings"

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

// allowedOrigin returns true if the origin is permitted.
func isOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	
	// 1. Allow Localhost
	if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
		return true
	}

	// 2. Allow Vercel subdomains
	if strings.HasSuffix(origin, ".vercel.app") {
		return true
	}

	// 3. Check environment variable
	if envOrigin := os.Getenv("ALLOWED_ORIGIN"); envOrigin != "" {
		for _, o := range strings.Split(envOrigin, ",") {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
	}

	// 4. Hardcoded fallbacks
	allowed := []string{
		"https://snapcheats-frontend.vercel.app",
		"https://snap-cheats-frontend.vercel.app",
	}
	for _, o := range allowed {
		if o == origin {
			return true
		}
	}
	return false
}

// Handler is the main entrypoint for Vercel serverless functions
func Handler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers to the response writer directly for preflights
	origin := r.Header.Get("Origin")
	if isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-API-Key")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Vary", "Origin")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	app.ServeHTTP(w, r)
}
