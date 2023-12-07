package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/subscription"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/sessions"
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

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("deleting user '%s'", user))
	filter := bson.M{"User": user}
	_, err = auth_db.DeleteOne(context.TODO(), filter)
	if err == nil {
		return true
	} else if err == mongo.ErrNoDocuments {
		logger.Warn("Unable to delete non-existent user")
		return false
	} else {
		logger.Error("Generic mongo delete error: ", err)
		return false
	}
}

func switchToCrypto(logger *logrus.Logger, user string) error {
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
		return err
	}
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	filter := bson.M{"User": user}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		logger.Error("Generic mongo read error: ", err)
		return err
	}

	//check if this is bogus
	if result["cryptoBilling"] != nil && result["cryptoBilling"].(bool) == true {
		logger.Warn("bogus crypto billing change attempt on user ", user)
	} else {
		logger.Debug("Updating user to crypto billing -> ", user)
		//else lets modify the document after updating stripe
		//stripe
		if result["cardBillingId"] == nil {
			logger.Warn("database cardBillingId doesnt exist for user ", user)
			return nil
		}

		_, err := subscription.Get(result["cardBillingId"].(string), nil)
		if err != nil {
			logger.Error("error finding cardBillingId in stripe for user ", user)
			return err
		}

		// Cancel the subscription
		params := &stripe.SubscriptionCancelParams{
			Params: stripe.Params{},
		}
		_, err = subscription.Cancel(result["cardBillingId"].(string), params)
		if err != nil {
			logger.Error("error canceling subscription in stripe for user ", user)
			return err
		}

		//db update
		update := bson.M{
			"$set": bson.M{
				"cryptoBilling":       true,
				"billingEmailSent":    false,
				"billingReminderSent": false,
				"billingToken":        generateToken(),
			},
			"$unset": bson.M{
				"cardBilling":   "",
				"cardBillingId": "",
			},
		}
		_, err = auth_db.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			logger.Error("Error setting user crypto billing data: ", err)
			return err
		}
		logger.Debug("UpdatED user to crypto billing -> ", user)

		//send email to the end user
		sendBillingEmail(logger, user)

		logger.Debug("sent crypto billing email -> ", user)
	}

	return nil
}

func switchToCard(logger *logrus.Logger, session *sessions.Session) error {
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
		return err
	}
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	user := session.Values["username"].(string)
	filter := bson.M{"User": user}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		logger.Error("Generic mongo read error: ", err)
		return err
	}

	//check the subscription we want to switch to
	if session.Values["billingId"] == nil {
		logger.Warn("session billingId doesnt exist for user ", user)
		return nil
	}

	//check the existing database subscription; cancel if exists
	if result["cardBillingId"] != nil {
		_, err := subscription.Get(result["cardBillingId"].(string), nil)
		if err != nil {
			logger.Error("error finding cardBillingId in stripe for user ", user)
			return err
		}

		// Cancel the subscription
		params := &stripe.SubscriptionCancelParams{
			Params: stripe.Params{},
		}
		_, err = subscription.Cancel(result["cardBillingId"].(string), params)
		if err != nil {
			logger.Error("error canceling subscription in stripe for user ", user)
			return err
		}
	}

	sub, err := subscription.Get(session.Values["billingId"].(string), nil)
	if err != nil {
		logger.Error("error finding cardBillingId in stripe for user ", user)
		return err
	}

	if sub.Status != `trialing` && sub.Status != `active` {
		logger.Warn(fmt.Sprintf("User %s is trying to set their account to inactive subscription", user))
		return nil
	}

	//db update with new subscription
	update := bson.M{
		"$set": bson.M{
			"cardBilling":   true,
			"cardBillingId": session.Values["billingId"].(string),
		},
		"$unset": bson.M{
			"cryptoBilling":       "",
			"billingEmailSent":    "",
			"billingReminderSent": "",
			"billingToken":        "",
		},
	}
	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user crypto billing data: ", err)
		return err
	}

	//send email to the end user
	sendBillingEmail(logger, user)

	return nil
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
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

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
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

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
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

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
	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

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

	var update bson.M
	//update the item to set the user to active
	if result["cryptoBilling"] != nil && result["cryptoBilling"] == true {
		update = bson.M{
			"$set": bson.M{
				"Active":           true,
				"billingEmailSent": true,
			},
		}
	} else {
		update = bson.M{
			"$set": bson.M{
				"Active": true,
			},
		}
	}

	_, err = auth_db.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error("Error setting user to active: ", err)
		return false
	}

	return true
}

func getUserData(logger *logrus.Logger, user string) (sessionData, error) {
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
		return sessionData{}, err
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
			return
		}
	}()

	auth_db := client.Database("auth").Collection("keys")
	filter := bson.M{"User": user}

	// Check if the item exists in the collection
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return sessionData{}, err
	}

	data := sessionData{
		Captcha:         false,
		Stripe:          false,
		Username:        result["User"].(string),
		Email:           result["Email"].(string),
		Active:          result["Active"].(bool),
		Crypto:          false,
		Card:            false,
		BillingCycleEnd: result["billingCycleEnd"].(string),
	}

	if result["cryptoBilling"] != nil {
		data.Crypto = true
		data.Token = result["billingToken"].(string)
	} else if result["cardBilling"] != nil {
		data.Card = true
	} else {
		logger.Warn("Unable to find billing type from db for: ", user)
	}

	return data, nil
}
