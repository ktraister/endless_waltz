package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

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

//to optimize reaper, spin off a goroutine to keep a slice updated and pull from that when writing
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

func insertItems(logger *logrus.Logger, ctx context.Context, count int64, otp_db *mongo.Collection) bool {
	// Create an array of documents to insert
	documents := []interface{}{}

	for i := 0; i < int(count); i++ {
		// create uuid inside function for ease of use
		id := uuid.New().String()
		ok := checkUUIDUnique(logger, ctx, otp_db, id)
		if !ok {
			logger.Debug("Non-unique UUID generated, passing for now...")
			continue
		}

		otp, err := createOTP()
		if err != nil {
			logger.Error(err)
			return false
		}

		//add our new OTP document to our array
		documents = append(documents, bson.D{{"UUID", id}, {"Pad", otp}})
	}

	// Create an array of insert models
	var insertModels []mongo.WriteModel
	for _, doc := range documents {
		insertModels = append(insertModels, mongo.NewInsertOneModel().SetDocument(doc))
	}

	//Then we insert
	result, err := otp_db.BulkWrite(ctx, insertModels)
	if err != nil {
		writeFailedCount++
		if writeFailedCount <= 10 {
			logger.Warn(err)
		} else {
			logger.Error(err)
		}
		return false
	}
	logger.Info(fmt.Sprintf("Inserted %d documents", result.InsertedCount))
	return true
}

func checkEntropy(logger *logrus.Logger) bool {
	return true
}

func main() {
	//reading in env variable for mongo conn URI
	MongoURI := os.Getenv("MongoURI")
	MongoUser := os.Getenv("MongoUser")
	MongoPass := os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")
	WriteThreshold := os.Getenv("WriteThreshold")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Reaper finished starting up!")

	for {
		//check entropy file before connecting to db
		var ok bool
		ok = checkEntropy(logger)
		if !ok {
			logger.Fatal("Entropy Heartbeat File did not pass check...")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		credential := options.Credential{
			Username: MongoUser,
			Password: MongoPass,
		}
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
		if err != nil {
			logger.Fatal(err)
		}

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			logger.Fatal(err)
		}
		logger.Info("Database connection succesful!")

		otp_db := client.Database("otp").Collection("otp")

		count := otpItemCount(logger, ctx, otp_db)
		if count == -1 {
			logger.Warn("Unable to count items in DB!")
			continue
		}

		threshold, err := strconv.ParseInt(WriteThreshold, 10, 64)
		if err != nil {
			threshold = 0
		}
		//if count is less than threshold
		if count != threshold {
			for count < threshold {
				diff := threshold - count
				if diff > 100 {
					diff = 100
				}
				logger.Info("Found count ", count, ", writing ", diff, " to db...")
				ok = insertItems(logger, ctx, diff, otp_db)
				if !ok {
					break
				}
				count = count + diff
			}
		} else {
			logger.Info("Count met threshold, sleeping...")
		}
		time.Sleep(10 * time.Second)
	}
}
