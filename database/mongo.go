package database

import (
	"log"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	ctx *context.Context
)


func init() {
	MongoClient = connectToMongoDB()
}

func DisconnectMongoClient() {
	MongoClient.Disconnect(*ctx)
}


func connectToMongoDB() *mongo.Client  {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the MongoDB server to verify the connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return mongoClient
}