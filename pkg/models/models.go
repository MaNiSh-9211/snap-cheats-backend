package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Admin struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"-"`
}

type DeviceInfo struct {
	Username string `bson:"username" json:"username"`
	Hostname string `bson:"hostname" json:"hostname"`
	Platform string `bson:"platform" json:"platform"`
	Arch     string `bson:"arch" json:"arch"`
	OSType   string `bson:"osType" json:"osType"`
	Release  string `bson:"release" json:"release"`
}

type Keylog struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeviceInfo     DeviceInfo         `bson:"deviceInfo" json:"deviceInfo"`
	QuestionNumber string             `bson:"questionNumber" json:"questionNumber"`
	LoggedKeys     string             `bson:"loggedKeys" json:"loggedKeys"`
	LastTimestamp  time.Time          `bson:"lastTimestamp" json:"lastTimestamp"`
}

type KeylogResponse struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	QuestionNumber string             `bson:"questionNumber" json:"questionNumber"`
	Response       string             `bson:"response" json:"response"`
	SubmittedAt    time.Time          `bson:"submittedAt" json:"submittedAt"`
	KeylogID       primitive.ObjectID `bson:"keylogId" json:"keylogId"`
	FirstPulledAt  *time.Time         `bson:"firstPulledAt,omitempty" json:"firstPulledAt,omitempty"`
	LastPulledAt   *time.Time          `bson:"lastPulledAt,omitempty" json:"lastPulledAt,omitempty"`
}

type Question struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	User        string               `bson:"user" json:"user"`
	Question    string               `bson:"question" json:"question"`
	Image       []byte               `bson:"image" json:"image"`
	ContentType string               `bson:"contentType" json:"contentType"`
	Responses   []primitive.ObjectID `bson:"responses" json:"responses"`
}

type ImageResponse struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	QuestionID     primitive.ObjectID `bson:"questionId" json:"questionId"`
	QuestionNumber string             `bson:"questionNumber" json:"questionNumber"`
	Response       string             `bson:"response" json:"response"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	FirstPulledAt  *time.Time         `bson:"firstPulledAt,omitempty" json:"firstPulledAt,omitempty"`
	LastPulledAt   *time.Time          `bson:"lastPulledAt,omitempty" json:"lastPulledAt,omitempty"`
}
