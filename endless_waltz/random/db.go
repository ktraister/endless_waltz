package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func listAllUsers(logger *logrus.Logger) (string, error) {
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
		logger.Error(err)
		return "", err
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	var filter, result bson.M
	// Perform the query
	cursor, err := auth_db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Resolve mongo find error", err)
		return "", err
	}
	defer cursor.Close(context.TODO())

	var final string
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Resolve mongo result decode error: ", err)
			continue
		}
		final = final + result["User"].(string) + ":"
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Resolve mongo cursor error: ", err)
	}

	return final, nil
}

func checkFriendsList(user string, logger *logrus.Logger) (string, error) {
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
		logger.Error(err)
		return "", err
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	var filter, result bson.M
	filter = bson.M{"User": user, "Active": true}
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	if result["FriendsList"] == nil {
		logger.Debug("you have no friends ha ha")
		return "", nil
	}

	return result["FriendsList"].(string), nil
}

func updateFriendsList(logger *logrus.Logger, user string, userList string) bool {
	logger.Debug(fmt.Sprintf("Updating friends list for %s with UserList %s", user, userList))

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
		logger.Error(err)
		return false
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	//first we grab the current friends list
	var result bson.M
	filter := bson.M{"User": user, "Active": true}
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		logger.Error(err)
		return false
	}

	update := bson.M{
		"$set": bson.M{
			"FriendsList": userList,
		},
	}

	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user crypto billing data: ", err)
		return false
	}

	return true
}

// checkAuth needs to get updated to allow users to login with deactive accts
func checkSub(user string, logger *logrus.Logger) string {
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
		logger.Error(err)
		return "false"
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	var filter, result bson.M
	filter = bson.M{"User": user, "Active": true}
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Debug("Found creds in db, authorized")
		if result["Premium"].(bool) == true {
			return "premium"
		} else {
			return "basic"
		}
	} else if err == mongo.ErrNoDocuments {
		logger.Info("No creds found, unauthorized")
		return "false"
	} else {
		logger.Error(err)
		return "fack"
	}
}
