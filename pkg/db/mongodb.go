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
		log.Fatal("MONGODB_URI not set in environment")
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
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
