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
	MongoUser := os.Getenv("MongoUser")
	MongoPass := os.Getenv("MongoPass")
        fmt.Println("MongoURI: ", MongoURI)
	fmt.Println("Reaper finished starting up!")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		credential := options.Credential{
		    Username: MongoUser,
		    Password: MongoPass,
		}
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
		if err != nil {
			panic(err)
		} else {
		    fmt.Println("Database connection succesful!")
		}

		otp_db := client.Database("otp").Collection("otp")
		b := make([]byte, 4096)

		//check and see how many items are in the db
		filter := bson.D{{}}
		count, err := otp_db.CountDocuments(ctx, filter)
		if err != nil {
			panic(err)
		}

		//if count is less than threshold
		if count < 100 {
		    fmt.Println("Found count ", count, "writing to db...")
			for i := 0; i < 100-int(count); i++ {
				//read from random
				_, err := rand.Read(b)
				id := uuid.New().String()
				_, err = otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"Pad", fmt.Sprintf("%v", b)}})
				if err != nil {
					fmt.Println(err)
				}

                                fmt.Println("Wrote item ", i, " to DB!")
			}
                    fmt.Println("Done writing to DB!")
		} 

		fmt.Println("Count met threshold, sleeping...")
		time.Sleep(10 * time.Second)
	}
}
