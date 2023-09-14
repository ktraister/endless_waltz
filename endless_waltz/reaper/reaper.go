package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

func main() {
	//reading in env variable for mongo conn URI
	MongoURI := os.Getenv("MongoURI")
	MongoUser := os.Getenv("MongoUser")
	MongoPass := os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Reaper finished starting up!")

	//set counts to use later in error message escalation
	readFailedCount := 0
	writeFailedCount := 0

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		credential := options.Credential{
			Username: MongoUser,
			Password: MongoPass,
		}
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
		if err != nil {
			logger.Fatal(err)
			return
		} else {
			logger.Info("Database connection succesful!")
		}
		otp_db := client.Database("otp").Collection("otp")

		b := make([]byte, 4096)

		//check and see how many items are in the db
		filter := bson.D{{}}
		count, err := otp_db.CountDocuments(ctx, filter)
		if err != nil {
			readFailedCount++
			if readFailedCount <= 2 {
				logger.Warn(err)
			} else {
				logger.Error(err)
			}
			break
		}

		//if count is less than threshold (this will need to go up for prod)
		threshold := int64(1000)
		if count < threshold {
			logger.Info("Found count ", count, ", writing to db...")
			for i := int64(0); i < threshold-count; i++ {
				//read from random
				_, err := rand.Read(b)
				id := uuid.New().String()
				//need to check if UUID already exists in db
				// Define the filter criteria
				filter := bson.M{"UUID": id} 

				// Check if the item exists in the collection
				var result bson.M
				err = otp_db.FindOne(context.TODO(), filter).Decode(&result)
				if err == nil {
					logger.Warn("UUID exists in the collection, passing.")
					continue
				} else if err == mongo.ErrNoDocuments {
					logger.Debug("UUID is unique, proceeding.")
				} else {
					logger.Error(err)
				}

				//Then we insert
				_, err = otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"Pad", fmt.Sprintf("%v", b)}})
				if err != nil {
					writeFailedCount++
					if writeFailedCount <= 10 {
						logger.Warn(err)
					} else {
						logger.Error(err)
					}
					continue
				}
				logger.Debug("Wrote item ", i, " to DB!")
			}
			logger.Info("Done writing to DB!")
		}

		logger.Info("Count met threshold, sleeping...")
		time.Sleep(10 * time.Second)
	}
}
