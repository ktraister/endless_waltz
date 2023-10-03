package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var MongoURI string
var MongoUser string
var MongoPass string

func deleteUser(logger *logrus.Logger, user string) bool {
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

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("deleting user '%s'", user))
	filter := bson.M{"User": user}
	var result bson.M
	//delete one
	err = auth_db.FinOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Info("Found creds in db, authorized")
		return true
	} else if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else {
		logger.Error(err)
		return false
	}
}

