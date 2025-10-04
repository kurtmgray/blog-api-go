package database

import (
	"context"
	"log"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
	Database *mongo.Database
}

func Connect(uri string) (*MongoDB, error) {
	// context, as req in Node
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	
	log.Println("Connected to MongoDB")

	// get database
	database := client.Database("blog")

	return &MongoDB{
		Client: client,
		Database: database,
	}, nil
}

func (db *MongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return db.Client.Disconnect(ctx)
}