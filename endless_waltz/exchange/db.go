package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func getPremiumUsers(logger *logrus.Logger) ([]string, error) {
	//creating context to connect to mongo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	credential := options.Credential{
		Username: MongoUser,
		Password: MongoPass,
	}
	//actually connect to mongo
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
	if err != nil {
		return []string{}, err
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")
	filter := bson.M{"Premium": true, "Active": true}
	tmpMap := []string{}

	// Set up a cursor to retrieve documents
	cursor, err := auth_db.Find(ctx, filter, options.Find())
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)

	// Iterate over the documents using the cursor
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return []string{}, err
		}
		tmpMap = append(tmpMap, result["User"].(string))
	}

	// Check for cursor errors after the loop
	if err := cursor.Err(); err != nil {
		return []string{}, err
	}

	return tmpMap, nil
}
