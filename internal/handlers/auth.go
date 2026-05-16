package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
)

// hashPassword computes the SHA-256 hash of a password string
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var admin models.Admin
	log.Printf("Login attempt for username: %s", input.Username)
	// Using bson.M for query mapping prevents NoSQL injection as the driver handles escaping
	err := db.KeyloggerDB.Collection("admins").FindOne(context.Background(), bson.M{"username": input.Username}).Decode(&admin)
	if err != nil {
		log.Printf("User %s not found in database", input.Username)
		// Generic error message to prevent username enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	log.Printf("User %s found, comparing hashes...", input.Username)
	// Compare SHA-256 hashed password
	if admin.Password != hashPassword(input.Password) {
		log.Printf("Password mismatch for user %s", input.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  admin.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
