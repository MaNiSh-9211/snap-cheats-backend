package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/handlers"
	"snap-monolith/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var app *gin.Engine

func init() {

	// Load .env locally only
	_ = godotenv.Load()

	// Connect database
	db.Connect()

	// Create Gin app
	app = gin.Default()

	// =====================================
	// Security Headers
	// =====================================
	app.Use(func(c *gin.Context) {

		c.Writer.Header().Set(
			"X-Content-Type-Options",
			"nosniff",
		)

		c.Writer.Header().Set(
			"X-Frame-Options",
			"DENY",
		)

		c.Writer.Header().Set(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		// Optional CSP
		c.Writer.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		)

		c.Next()
	})

	// =====================================
	// CORS Configuration
	// =====================================
	app.Use(cors.New(cors.Config{

		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"https://snapcheats-frontend.vercel.app",
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept",
			"Authorization",
			"X-API-Key",
			"X-Requested-With",
		},

		ExposeHeaders: []string{
			"Content-Length",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	// =====================================
	// Handle OPTIONS Requests
	// =====================================
	app.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// =====================================
	// Public Routes
	// =====================================
	app.POST("/api/login", handlers.Login)

	// =====================================
	// App Routes
	// =====================================
	appGroup := app.Group("/api")
	appGroup.Use(middleware.AppApiKeyMiddleware())

	{
		appGroup.POST(
			"/text/sync",
			handlers.SyncKeylog,
		)

		appGroup.GET(
			"/text/response/:questionNumber",
			handlers.GetKeylogResponse,
		)

		appGroup.POST(
			"/image/upload",
			handlers.UploadQuestion,
		)

		appGroup.GET(
			"/image/response/:questionNumber",
			handlers.GetImageResponse,
		)
	}

	// =====================================
	// Admin Routes
	// =====================================
	admin := app.Group("/api")
	admin.Use(middleware.AuthMiddleware())

	{
		// Text Routes
		admin.GET(
			"/text/logs",
			handlers.GetAllKeylogs,
		)

		admin.GET(
			"/text/responses/id/:keylogId",
			handlers.GetKeylogResponsesByID,
		)

		admin.POST(
			"/text/responses",
			handlers.SubmitKeylogResponse,
		)

		admin.DELETE(
			"/text/logs/:id",
			handlers.DeleteKeylog,
		)

		admin.DELETE(
			"/text/responses/all/:keylogId",
			handlers.DeleteKeylogResponses,
		)

		admin.DELETE(
			"/text/responses/:id",
			handlers.DeleteKeylogResponse,
		)

		// Image Routes
		admin.GET(
			"/image/questions",
			handlers.GetAllQuestions,
		)

		admin.GET(
			"/image/response/id/:questionId",
			handlers.GetImageResponsesByID,
		)

		admin.POST(
			"/image/response",
			handlers.SubmitResponse,
		)

		admin.DELETE(
			"/image/questions/:id",
			handlers.DeleteImage,
		)

		admin.DELETE(
			"/image/responses/:id",
			handlers.DeleteImageResponse,
		)
	}
}

// =====================================
// Vercel Serverless Handler
// =====================================
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

// =====================================
// Local Development Server
// =====================================
func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)

	if err := app.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}