package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var KeyloggerDB *mongo.Database
var AutoCheatDB *mongo.Database

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("MONGODB_URI not set in environment - database disabled")
		return
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Println("mongo.Connect failed:", err)
		return
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("MongoDB ping failed (check Atlas IP whitelist 0.0.0.0/0 and credentials):", err)
	} else {
		log.Println("Connected to MongoDB")
	}

	Client = client

	keyloggerDBName := os.Getenv("KEYLOGGER_DB_NAME")
	if keyloggerDBName == "" {
		keyloggerDBName = "keylogger"
	}
	KeyloggerDB = client.Database(keyloggerDBName)

	autoCheatDBName := os.Getenv("AUTOCHEAT_DB_NAME")
	if autoCheatDBName == "" {
		autoCheatDBName = "autocheat"
	}
	AutoCheatDB = client.Database(autoCheatDBName)

	log.Println("Connected to MongoDB")
}
