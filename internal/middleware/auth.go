package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Allow preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})

			c.Abort()
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")

		if jwtSecret == "" {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "JWT_SECRET environment variable missing",
			})

			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			// Ensure signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}

func AppApiKeyMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Allow preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")

		expectedKey := os.Getenv("APP_API_KEY")

		// Allow if APP_API_KEY is not configured
		if expectedKey == "" {
			c.Next()
			return
		}

		if apiKey != expectedKey {

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid API Key",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}