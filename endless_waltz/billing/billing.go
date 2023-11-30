package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

var MongoURI, MongoUser, MongoPass string

func cryptoResolvePayments(logger *logrus.Logger) {
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
		logger.Error("Resolve mongo connect error: ", err)
		return
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	db := client.Database("auth").Collection("keys")

	//find all records where billingCharge != nil
	filter := bson.D{
		{"item", bson.D{
			{"$exists", true},
		}},
	}

	// Perform the query
	cursor, err := db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Resolve mongo find error", err)
		return
	}
	defer cursor.Close(context.TODO())

	index := 0
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		index += 1
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Resolve mongo result decode error: ", err)
			continue
		}

		//check if the charge was payed
		url := fmt.Sprintf("https://api.commerce.coinbase.com/charges/%s", result["billingCharge"])
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logger.Error("Resolve error creating request: ", err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-CC-Api-Key", os.Getenv("CoinbaseAPIKey"))

		res, err := client.Do(req)
		if err != nil {
			logger.Error("Resolve error doing req: ", err)
			continue
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error("Resolve error reading response: ", err)
			continue
		}
		fmt.Println(string(body))

		/*
			if paid {
				//set billingCharge nil
				updateFilter := bson.M{"_id": result["_id"]}
				update := bson.M{
					"$set": bson.M{
						"billingEmailSent": true,
					},
				}
				_, err = db.UpdateOne(ctx, updateFilter, update)
				if err != nil {
					logger.Error("Resolve mongo update error: ", err)
				}
			}
		*/
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Resolve mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Resolve Updated %d records", index))
}

func cryptoBillingInit(logger *logrus.Logger) {
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
		logger.Error("Init mongo connect error: ", err)
		return
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	db := client.Database("auth").Collection("keys")

	today := time.Now()
	threshold := today.Add(168 * time.Hour).Format("01-02-2006")
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": false, "billingCyclePaid": false, "billingCycleEnd": bson.M{"$lte": threshold}}

	// Perform the query
	cursor, err := db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Init mongo find error", err)
		return
	}
	defer cursor.Close(context.TODO())

	index := 0
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		index += 1
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Init mongo result decode error: ", err)
			continue
		}

		//set billingEmailSent true
		updateFilter := bson.M{"_id": result["_id"]}
		update := bson.M{
			"$set": bson.M{
				"billingEmailSent": true,
			},
		}
		_, err = db.UpdateOne(ctx, updateFilter, update)
		if err != nil {
			logger.Error("Init mongo update error: ", err)
			continue
		}

		//send the INIT billing email
		sendCryptoBillingEmail(logger, result["User"].(string), result["Email"].(string), result["billingToken"].(string))
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Init mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Init Updated %d records", index))
}

func cryptoBillingReminder(logger *logrus.Logger) {
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
		logger.Error("Reminder mongo connect error: ", err)
		return
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	db := client.Database("auth").Collection("keys")

	today := time.Now()
	threshold := today.Add(48 * time.Hour).Format("01-02-2006")
	//find all records where Active:true, cryptoBilling:true, billingEmailSent: true, billingReminderSent, false, billingCyclePaid:false
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": true, "billingReminderSent": false, "billingCyclePaid": false, "billingCycleEnd": bson.M{"$lte": threshold}}

	// Perform the query
	cursor, err := db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Reminder mongo find error", err)
		return
	}
	defer cursor.Close(context.TODO())

	index := 0
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		index += 1
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Reminder mongo result decode error: ", err)
			continue
		}

		//set billingReminderSent true
		updateFilter := bson.M{"_id": result["_id"]}
		update := bson.M{
			"$set": bson.M{
				"billingReminderSent": true,
			},
		}
		_, err = db.UpdateOne(ctx, updateFilter, update)
		if err != nil {
			logger.Error("Reminder mongo update error: ", err)
			continue
		}

		//send the reminder billing email
		sendCryptoBillingReminder(logger, result["User"].(string), result["Email"].(string), result["billingToken"].(string))
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Reminder mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Reminder Updated %d records", index))
}

func cryptoDisableAccount(logger *logrus.Logger) {
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
		logger.Error("Disable mongo connect error: ", err)
		return
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	db := client.Database("auth").Collection("keys")

	today := time.Now().Format("01-02-2006")
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": true, "billingReminderSent": true, "billingCyclePaid": false, "billingCycleEnd": bson.M{"$lt": today}}

	// Perform the query
	cursor, err := db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Disable mongo find error", err)
		return
	}
	defer cursor.Close(context.TODO())

	index := 0
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		index += 1
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Disable mongo result decode error: ", err)
			continue
		}

		//disable account
		updateFilter := bson.M{"_id": result["_id"]}
		update := bson.M{
			"$set": bson.M{
				"Active": false,
			},
		}
		_, err = db.UpdateOne(ctx, updateFilter, update)
		if err != nil {
			logger.Error("Disable mongo update error: ", err)
			continue
		}

		//send disable email
		sendCryptoBillingDisabled(logger, result["User"].(string), result["Email"].(string), result["billingToken"].(string))
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Disable mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Disable Updated %d records", index))
}

// this could just be a cron job that runs daily...
func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Billing binary finished starting up!")

	//crypto billing
	//crypto resolve payments
	cryptoResolvePayments(logger)

	//crypto billing init (7 days before expire)
	cryptoBillingInit(logger)

	//crypto billing reminder (2 days before expire)
	cryptoBillingReminder(logger)

	//crypto billing disable after cycle end
	cryptoDisableAccount(logger)
}
