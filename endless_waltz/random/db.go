package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

// checkAuth needs to get updated to allow users to login with deactive accts
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

	return result["FriendsList"].(string), nil
}

func updateFriendsList(logger *logrus.Logger, user string, targetUser string, action string) bool {
	logger.Debug(fmt.Sprintf("Updating friends list for %s, action %s with user %s", user, action, targetUser))

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

	var inList, nilList bool
	if result["FriendsList"] == nil {
		inList = false
		nilList = true
	} else {
		inList = strings.Contains(result["FriendsList"].(string), targetUser)
		nilList = false
	}
	if action == "Add" {
		if inList {
			logger.Warn(fmt.Sprintf("Bogus user addToFriendsList for user %s => %s", user, targetUser))
			return true
		}

		var update bson.M
		if !nilList {
			update = bson.M{
				"$set": bson.M{
					"FriendsList": strings.Replace(fmt.Sprintf(result["FriendsList"].(string) + ":" + targetUser), "::", ":", -1),
				},
			}
		} else {
			update = bson.M{
				"$set": bson.M{
					"FriendsList": targetUser + ":",
				},
			}

		}
		_, err = auth_db.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			logger.Error("Error setting user crypto billing data: ", err)
			return false
		}
	} else if action == "Remove" {
		if !inList || nilList {
			logger.Warn(fmt.Sprintf("Bogus user removeFromFriendsList for user %s => %s", user, targetUser))
			return true
		} 

		tmp := strings.Replace(result["FriendsList"].(string), targetUser, "", -1)
		tmp = strings.Replace(tmp, "::", ":", -1)

		//db update
		update := bson.M{
			"$set": bson.M{
				"FriendsList": tmp,
			},
		}
		_, err = auth_db.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			logger.Error("Error setting user crypto billing data: ", err)
			return false
		}
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
