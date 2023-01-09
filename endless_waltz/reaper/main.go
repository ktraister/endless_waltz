package main

import (
        "context"
	"crypto/rand"
	"fmt"
	"time"
	"os"

        "go.mongodb.org/mongo-driver/bson"
        "go.mongodb.org/mongo-driver/mongo"
        "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

func main() {
        //reading in env variable for mongo conn URI
	MongoURI := os.Getenv("MongoURI")
        fmt.Println("MongoURI: ", MongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		fmt.Println(err)
	}

	otp_db := client.Database("otp").Collection("otp")
	b := make([]byte, 4096)

	for {
		//check and see how many items are in the db
		filter := bson.D{{}}
		count, err := otp_db.CountDocuments(ctx, filter)
		if err != nil {
			panic(err)
		}

		//if count is less than threshold
		if count < 100 {
			for i := 0; i < 100-int(count); i++ {
				//read from random
				n, err := rand.Read(b)
				fmt.Println(n, err, b)

				//uuid create and send to db
				id := uuid.New().String()
				_, err = otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"Pad", fmt.Sprintf("%v", b)}})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
