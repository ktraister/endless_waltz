package main

import (
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// set counts to use later in error message escalation
var readFailedCount = 0
var writeFailedCount = 0
var result bson.M
var b = make([]byte, 4096)

func otpItemCount(logger *logrus.Logger, ctx context.Context, otp_db *mongo.Collection) int64 {
	filter := bson.D{{}}
	count, err := otp_db.CountDocuments(ctx, filter)
	if err != nil {
		readFailedCount++
		if readFailedCount <= 2 {
			logger.Warn(err)
		} else {
			logger.Error(err)
		}
		return -1
	}
	return int64(count)
}

func checkUUIDUnique(logger *logrus.Logger, ctx context.Context, otp_db *mongo.Collection, id string) bool {
	//need to check if UUID already exists in db
	// Define the filter criteria
	filter := bson.M{"UUID": id}

	// Check if the item exists in the collection
	err := otp_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Warn("UUID exists in the collection, passing.")
		return false
	} else if err == mongo.ErrNoDocuments {
		logger.Debug("UUID is unique, proceeding.")
		return true
	} else {
		logger.Error(err)
		return false
	}
}

func createOTP() (string, error) {
	temp := []string{}
	maximum, _ := big.NewInt(0).SetString("1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", 0)
	i := 0
	for i < 4097 {
		randomNumber, _ := rand.Int(rand.Reader, maximum)
		temp = append(temp, randomNumber.String())
		i++
	}
	return strings.Join(temp[:], " "), nil
}

func insertItem(logger *logrus.Logger, ctx context.Context, otp_db *mongo.Collection) bool {
	// create uuid inside function for ease of use
	id := uuid.New().String()
	ok := checkUUIDUnique(logger, ctx, otp_db, id)
	if !ok {
		logger.Debug("Non-unique UUID generated, passing for now...")
		return false
	}

	otp, err := createOTP()
	if err != nil {
		logger.Error(err)
		return false
	}

	//Then we insert
	_, err = otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"Pad", otp}})
	if err != nil {
		writeFailedCount++
		if writeFailedCount <= 10 {
			logger.Warn(err)
		} else {
			logger.Error(err)
		}
		return false
	}
	return true
}

func main() {
	//reading in env variable for mongo conn URI
	MongoURI := os.Getenv("MongoURI")
	MongoUser := os.Getenv("MongoUser")
	MongoPass := os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Reaper finished starting up!")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

		count := otpItemCount(logger, ctx, otp_db)
		if count == -1 {
			continue
		}

		//if count is less than threshold (this will need to go up for prod)
		threshold := int64(1000)
		if count < threshold {
			logger.Info("Found count ", count, ", writing to db...")
			for i := int64(0); i < threshold-count; i++ {
				ok := insertItem(logger, ctx, otp_db)
				if !ok {
					break
				} else {
					logger.Debug("Wrote item ", i, " to DB!")
				}
			}
			logger.Info("Done writing to DB!")
		}

		logger.Info("Count met threshold, sleeping...")
		time.Sleep(10 * time.Second)
	}
}
