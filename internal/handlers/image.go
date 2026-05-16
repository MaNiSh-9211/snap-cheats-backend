package handlers

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
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

func UploadQuestion(c *gin.Context) {
	var input struct {
		User        string `json:"user"`
		Question    string `json:"question"`
		Image       string `json:"image"` // Received as base64 string
		ContentType string `json:"contentType"`
	}

	var imgBytes []byte

	// Try JSON first
	if err := c.ShouldBindJSON(&input); err == nil {
		if len(input.Image) > 100 {
			log.Printf("Received Image (first 100 chars): %s", input.Image[:100])
		} else {
			log.Printf("Received Image: %s", input.Image)
		}
		// Manual decode for robustness
		decoded, err := base64.StdEncoding.DecodeString(input.Image)
		if err != nil {
			log.Printf("Base64 Decode Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Base64 decode failed: " + err.Error()})
			return
		}
		imgBytes = decoded
	} else {
		log.Printf("JSON Binding Error: %v", err)
		// Fallback to Multipart Form
		user := c.PostForm("user")
		question := c.PostForm("question")
		file, err := c.FormFile("image")
		if err == nil {
			f, _ := file.Open()
			defer f.Close()
			readBytes, _ := ioutil.ReadAll(f)
			input.User = user
			input.Question = question
			imgBytes = readBytes
			input.ContentType = file.Header.Get("Content-Type")
		} else {
			log.Printf("Multipart Binding Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format error: " + err.Error()})
			return
		}
	}

	if len(imgBytes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image data is empty"})
		return
	}

	q := models.Question{
		ID:          primitive.NewObjectID(),
		User:        input.User,
		Question:    input.Question,
		Image:       imgBytes,
		ContentType: input.ContentType,
		Responses:   []primitive.ObjectID{},
	}

	_, err := db.AutoCheatDB.Collection("images").InsertOne(context.Background(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question uploaded", "id": q.ID})
}

func DeleteImageResponse(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Find the response to get the questionId
	var resp models.ImageResponse
	err = db.AutoCheatDB.Collection("response").FindOne(context.Background(), bson.M{"_id": id}).Decode(&resp)
	if err == nil {
		// Pull from question's responses array
		_, _ = db.AutoCheatDB.Collection("images").UpdateOne(
			context.Background(),
			bson.M{"_id": resp.QuestionID},
			bson.M{"$pull": bson.M{"responses": id}},
		)
	}

	_, err = db.AutoCheatDB.Collection("response").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response deleted successfully"})
}

func GetImageResponse(c *gin.Context) {
	qNum := c.Param("questionNumber")
	
	var resp models.ImageResponse
	opts := options.FindOne().SetSort(bson.M{"createdAt": -1})
	err := db.AutoCheatDB.Collection("response").FindOne(context.Background(), bson.M{"questionNumber": qNum}, opts).Decode(&resp)
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
	_, _ = db.AutoCheatDB.Collection("response").UpdateOne(context.Background(), bson.M{"_id": resp.ID}, update)

	c.JSON(http.StatusOK, resp)
}

func GetImageResponsesByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("questionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	cursor, err := db.AutoCheatDB.Collection("response").Find(context.Background(), bson.M{"questionId": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var responses []models.ImageResponse
	if err = cursor.All(context.Background(), &responses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, responses)
}

func GetAllQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6")) // 6 items per page
	skip := (page - 1) * limit

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(skip))
	// No sort for now, or add if needed
	
	cursor, err := db.AutoCheatDB.Collection("images").Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.Background())

	var questions []models.Question
	if err = cursor.All(context.Background(), &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, questions)
}

func SubmitResponse(c *gin.Context) {
	var resp models.ImageResponse
	if err := c.ShouldBindJSON(&resp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp.ID = primitive.NewObjectID()
	resp.CreatedAt = time.Now()

	_, err := db.AutoCheatDB.Collection("response").InsertOne(context.Background(), resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update question responses array
	_, _ = db.AutoCheatDB.Collection("images").UpdateOne(
		context.Background(),
		bson.M{"_id": resp.QuestionID},
		bson.M{"$push": bson.M{"responses": resp.ID}},
	)

	c.JSON(http.StatusOK, gin.H{"message": "Response submitted", "id": resp.ID})
}

func DeleteImage(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_, err = db.AutoCheatDB.Collection("images").DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Optionally delete associated responses
	_, _ = db.AutoCheatDB.Collection("response").DeleteMany(context.Background(), bson.M{"questionId": id})

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}
