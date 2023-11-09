package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

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
		logger.Error("Could not connect to mongo:", err)
		return false
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("deleting user '%s'", user))
	filter := bson.M{"User": user}
	_, err = auth_db.DeleteOne(context.TODO(), filter)
	if err == nil {
		return true
	} else if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else {
		logger.Error("Generic mongo delete error: ", err)
		return false
	}
}

func prepareUserPassReset(logger *logrus.Logger, user string, token string) (string, error) {
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
		logger.Error("Could not connect to mongo:", err)
		return "", err
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("checking email for user '%s'", user))
	filter := bson.M{"User": user}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		logger.Error("Generic mongo read error: ", err)
		return "", err
	}

	//if the user exists, let's set their reset token and time
	update := bson.D{{"$set", bson.D{{"passwordResetToken", token}, {"passwordResetTime", time.Now().Unix()}}}}
	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user password reset data: ", err)
		return "", err
	}

	return result["Email"].(string), nil
}

func verifyPasswordReset(logger *logrus.Logger, email string, user string, token string) bool {
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
		logger.Error("verifyPasswordReset mongo generic error ", err)
		return false
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("resetting user pass for '%s'", user))
	filter := bson.M{"User": user, "Email": email, "passwordResetToken": token}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else if err != nil {
		logger.Error("Generic mongo FindOne error: ", err)
		return false
	}

	regTime := int(result["passwordResetTime"].(int64))

	threshold := regTime + 600

	//check the result for appropriate values
	if int(time.Now().Unix()) > threshold {
		logger.Warn(fmt.Sprintf("User '%s' exceeded password reset threshold", user))
		return false
	}

	return true
}

func submitPasswordReset(logger *logrus.Logger, email string, user string, token string, passHash string) bool {
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
		logger.Error("verifyUserSignup mongo generic error ", err)
		return false
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("resetting user pass for '%s'", user))
	filter := bson.M{"User": user, "Email": email, "passwordResetToken": token}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else if err != nil {
		logger.Error("Generic mongo FindOne error: ", err)
		return false
	}

	regTime := int(result["passwordResetTime"].(int64))

	threshold := regTime + 600

	//check the result for appropriate values
	if int(time.Now().Unix()) > threshold {
		logger.Warn(fmt.Sprintf("User '%s' exceeded password reset threshold", user))
		return false
	}

	//if the user exists, let's set their reset token and time
	update := bson.D{{"$set", bson.D{{"Password", passHash}, {"passwordResetToken", nil}, {"passwordResetTime", nil}}}}
	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user password reset data: ", err)
		return false
	}

	return true
}

func verifyUserSignup(logger *logrus.Logger, email string, user string, token string) bool {
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
		logger.Error("verifyUserSignup mongo generic error ", err)
		return false
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("verifying user '%s' w/ email '%s' and token '%s'", user, email, token))
	filter := bson.M{"User": user, "Email": email, "EmailVerifyToken": token}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else if err != nil {
		logger.Error("Generic mongo FindOne error: ", err)
		return false
	}

	regTime, err := strconv.Atoi(result["SignupTime"].(string))
	if err != nil {
		logger.Error("strconv error: ", err)
		return false
	}

	threshold := regTime + 600

	//check the result for appropriate values
	if int(time.Now().Unix()) > threshold {
		logger.Warn(fmt.Sprintf("User '%s' exceeded email verify threshold", user))
		return false
	}

	//update the item to set the user to active
	update := bson.D{{"$set", bson.D{{"Active", true}}}}
	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user to active: ", err)
		return false
	}

	return true
}
