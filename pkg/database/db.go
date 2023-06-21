package database

import (
	"context"
	"log"
	"time"

	"github.com/Rayato159/awaken-discord-bot/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func DBConnect(cfg config.IConfig) *mongo.Client {
	// Set time out for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to mongoDb
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Db().Url()))
	if err != nil || client == nil {
		log.Fatalf("connect to mongodb failed: %v", err)
	}
	// Ping to test connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatalf("ping to mongodb failed: %v", err)
	}
	defer log.Println("🍃mongodb has been connected, have a nice day!")
	return client
}

func MongoDbDisconnect(client *mongo.Client) {
	ctx := context.Background()
	log.Println("🍃mongodb has been disconnected, good bye")

	// Disconnect from database
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
