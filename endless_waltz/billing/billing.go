package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		{"billingCharge", bson.D{
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
	mod := 0
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

		var cbResp map[string]interface{}
		json.Unmarshal(body, &cbResp)
		//need to check here if the api returned an error
		if cbResp["error"] != nil {
			logger.Error("Error from coinbase API: ", cbResp["error"])
			return
		}

		timeline := cbResp["data"].(map[string]interface{})["timeline"].([]interface{})
		paid := false
		for _, item := range timeline {
			status := item.(map[string]interface{})["status"]
			if status == "COMPLETED" {
				paid = true
				break
			}
		}

		if paid {
			mod += 1
			//set billingCharge nil
			updateFilter := bson.M{"_id": result["_id"]}
			update := bson.M{
				"$set": bson.M{
				        "Active":              true,
					"billingEmailSent":    false,
					"billingReminderSent": false,
					"billingCycleEnd":     nextBillingCycle(result["billingCycleEnd"].(string)),
					"billingToken":        generateToken(),
				},
				"$unset": bson.M{
					"billingCharge": "",
				},
			}
			_, err = db.UpdateOne(ctx, updateFilter, update)
			if err != nil {
				logger.Error("Resolve mongo update error: ", err)
			}
			sendCryptoBillingThanks(logger, result["User"].(string), result["Email"].(string))
		}
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Resolve mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Resolve inspected %d records, updated %d", index, mod))
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
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": false, "billingCycleEnd": bson.M{"$lte": threshold}}

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
	//find all records where Active:true, cryptoBilling:true, billingEmailSent: true, billingReminderSent, false,
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": true, "billingReminderSent": false, "billingCycleEnd": bson.M{"$lte": threshold}}

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
	filter := bson.M{"Active": true, "cryptoBilling": true, "billingEmailSent": true, "billingReminderSent": true, "billingCycleEnd": bson.M{"$lt": today}}

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
		sendBillingDisabled(logger, "crypto", result["User"].(string), result["Email"].(string), result["billingToken"].(string))
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Disable mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("Disable Updated %d records", index))
}

func stripeSubscriptionChecks(logger *logrus.Logger) {
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

	//find all records where cardBilling = true
	filter := bson.M{
		"cardBilling": true,
	}

	// Perform the query
	cursor, err := db.Find(context.TODO(), filter)
	if err != nil {
		logger.Error("Disable mongo find error", err)
		return
	}
	defer cursor.Close(context.TODO())

	index := 0
	mod := 0
	// Iterate over the result records
	for cursor.Next(context.TODO()) {
		index += 1
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			logger.Error("Disable mongo result decode error: ", err)
			continue
		}

		if result["cardBillingId"] == nil {
			logger.Warn(fmt.Sprintf("Mongo document for %s has no subscription ID", result["User"].(string)))
			continue
		}

		params := &stripe.SubscriptionParams{}
		sub, err := subscription.Get(result["cardBillingId"].(string), params)
		if err != nil {
			logger.Error("Error getting stripe subscription data: ", err)
			continue
		}

		//stripe will keep track of billing cycles and all for us (!!!)
		//Possible values are `incomplete`, `incomplete_expired`, `trialing`, `active`, `past_due`, `canceled`, or `unpaid`.
		logger.Debug(fmt.Sprintf("checking sub %s for user %s, status --> %s", result["cardBillingId"], result["User"], sub.Status))
		if sub.Status != `trialing` && sub.Status != `active` {
			//account should be disabled
			if result["Active"] == true {
				mod += 1
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
				sendBillingDisabled(logger, "card", result["User"].(string), result["Email"].(string), "")
			}
		} else {
			//account should be enabled
			if result["Active"] == false {
				mod += 1
				//disable account
				updateFilter := bson.M{"_id": result["_id"]}
				update := bson.M{
					"$set": bson.M{
						"Active": true,
					},
				}
				_, err = db.UpdateOne(ctx, updateFilter, update)
				if err != nil {
					logger.Error("Disable mongo update error: ", err)
					continue
				}
			}
		}
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Disable mongo cursor error: ", err)
	}

	logger.Info(fmt.Sprintf("StripeChecks inspected %d records, updated %d records", index, mod))
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

	stripe.Key = os.Getenv("StripeAPIKey")

	//crypto billing
	//crypto resolve payments
	cryptoResolvePayments(logger)

	//crypto billing init (7 days before expire)
	cryptoBillingInit(logger)

	//crypto billing reminder (2 days before expire)
	cryptoBillingReminder(logger)

	//crypto billing disable after cycle end
	cryptoDisableAccount(logger)

	//check if card-billed accounts should be locked
	stripeSubscriptionChecks(logger)
}
