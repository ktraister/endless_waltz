package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

type User struct {
	Username string `json:"username"`
}

// primarily used for authentication and to test system health
func healthHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.Header.Get("User"), 5)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	ok = checkAuth(req.Header.Get("User"), req.Header.Get("Passwd"), true, logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	w.Write([]byte("HEALTHY"))
	logger.Info("Someone hit the health check route...")
}

// path to tell the  client if the user is basic or premium
func premiumHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.Header.Get("User"), 1)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	status := checkSub(req.Header.Get("User"), req.Header.Get("Passwd"), logger)
	if status == "premium" {
		w.Write([]byte("premium"))
	} else if status == "basic" {
		w.Write([]byte("basic"))
	} else {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

func cryptoPaymentHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.FormValue("user"), 1)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	//do a db lookup based on inputs
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
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
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

	// Check if the item exists in the collection
	// we're using the billingToken as a OTP
	user := req.FormValue("user")
	email := req.FormValue("email")
	token := req.FormValue("token")
	filter := bson.M{"User": user, "Email": email, "billingToken": token}
	var result bson.M
	err = db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Debug("Found creds in db, authorized")
	} else if err == mongo.ErrNoDocuments {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	} else {
		logger.Error(err)
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		return
	}

	domain := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		domain = "https://localhost"
	}

	//create the billing charge
	//https://docs.cloud.coinbase.com/commerce/docs/accepting-crypto#creating-a-charge
	//https://docs.cloud.coinbase.com/commerce/reference/createcharge
	payload := strings.NewReader(fmt.Sprintf(`{"name":"Endless Waltz Monthly Payment","redirect_url":"%s","pricing_type":"fixed_price","local_price":{"amount":"2.99","currency":"USD"}}`, domain))
	cReq, err := http.NewRequest("POST", "https://api.commerce.coinbase.com/charges", payload)
	if err != nil {
		logger.Error("Error creating billing charge: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}

	cReq.Header.Add("Content-Type", "application/json")
	cReq.Header.Add("Accept", "application/json")
	cReq.Header.Add("X-CC-Api-Key", os.Getenv("CoinbaseAPIKey"))
	httpClient := &http.Client{}
	res, err := httpClient.Do(cReq)
	if err != nil {
		logger.Error("Error performing https request to coinbase: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error("Disable mongo update error: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}
	var cbResp map[string]interface{}
	json.Unmarshal(body, &cbResp)

	//need to check here if the api returned an error
	if cbResp["error"] != nil {
		logger.Error("Error from coinbase API: ", cbResp["error"])
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}

	//update the billing charge in the db
	updateFilter := bson.M{"_id": result["_id"]}
	update := bson.M{
		"$set": bson.M{
			"billingCharge": cbResp["data"].(map[string]interface{})["code"],
		},
	}
	_, err = db.UpdateOne(ctx, updateFilter, update)
	if err != nil {
		logger.Error("Disable mongo update error: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}

	//redirect the user
	http.Redirect(w, req, cbResp["data"].(map[string]interface{})["hosted_url"].(string), http.StatusSeeOther)
}

func createCheckoutSession(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	domain := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		domain = "https://localhost"
	}

	params := &stripe.CheckoutSessionParams{
		UIMode:    stripe.String("embedded"),
		ReturnURL: stripe.String(domain + "/register?session_id={CHECKOUT_SESSION_ID}"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				// Provide the exact Price ID (for example, pr_1234) of the product you want to sell
				Price:    stripe.String("price_1OJe1UGcdL8YMSExsJZxn1J1"),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(30), // Set the trial period in days
		},
	}

	s, err := session.New(params)
	if err != nil {
		logger.Error("session.New: ", err)
	}

	writeJSON(w, struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: s.ClientSecret,
	}, logger)
}

func modifyCheckoutSession(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	domain := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		domain = "https://localhost"
	}

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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	var u User
	err = json.Unmarshal(body, &u)
	if err != nil {
		http.Error(w, "Error parsing JSON payload", http.StatusBadRequest)
		return
	}

	// WE NEED TO DO SOMETHING TO PASS A USER TO THIS FUNCTION
	// add params to post in json :)
	user := u.Username
	logger.Debug("Incoming user ", user)

	// Check if the item exists in the collection
	filter := bson.M{"User": user}
	var result bson.M
	err = db.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Database findOne error: ", err)
		return
	}

	layout := "01-02-2006"
	var daysLeft int64
	if result["billingCycleEnd"] != nil {
		dateString := result["billingCycleEnd"].(string)
		billingDate, _ := time.Parse(layout, dateString)
		today := time.Now()
		duration := billingDate.Sub(today)
		daysLeft = int64(duration.Hours()/24 + 1)
	} else {
		daysLeft = 0
	}

	var params *stripe.CheckoutSessionParams
	if daysLeft >= 1 {
		params = &stripe.CheckoutSessionParams{
			UIMode:    stripe.String("embedded"),
			ReturnURL: stripe.String(domain + "/switchToCard?session_id={CHECKOUT_SESSION_ID}"),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				&stripe.CheckoutSessionLineItemParams{
					// Provide the exact Price ID (for example, pr_1234) of the product you want to sell
					Price:    stripe.String("price_1OJe1UGcdL8YMSExsJZxn1J1"),
					Quantity: stripe.Int64(1),
				},
			},
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				TrialPeriodDays: stripe.Int64(daysLeft), // Set the trial period in days
			},
		}
	} else {
		params = &stripe.CheckoutSessionParams{
			UIMode:    stripe.String("embedded"),
			ReturnURL: stripe.String(domain + "/switchToCard?session_id={CHECKOUT_SESSION_ID}"),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				&stripe.CheckoutSessionLineItemParams{
					// Provide the exact Price ID (for example, pr_1234) of the product you want to sell
					Price:    stripe.String("price_1OJe1UGcdL8YMSExsJZxn1J1"),
					Quantity: stripe.Int64(1),
				},
			},
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		}
	}

	s, err := session.New(params)
	if err != nil {
		logger.Error("session.New: ", err)
	}

	writeJSON(w, struct {
		ClientSecret string `json:"clientSecret"`
	}{
		ClientSecret: s.ClientSecret,
	}, logger)
}

func retrieveCheckoutSession(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	session_id := req.URL.Query().Get("session_id")
	if session_id == "null" {
		logger.Warn("Null session id, returning {}")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	} else {
		//we grab the session details using the stripe SDK
		s, _ := session.Get(session_id, nil)

		writeJSON(w, struct {
			Status string `json:"status"`
		}{
			Status: string(s.Status),
		}, logger)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}, logger *logrus.Logger) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("json.NewEncoder.Encode: ", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		logger.Error("io.Copy: ", err)
		return
	}
}

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")
	stripe.Key = os.Getenv("StripeAPIKey")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Random Server finished starting up!")

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/api/healthcheck", healthHandler).Methods("GET")
	router.HandleFunc("/api/premiumCheck", premiumHandler).Methods("GET")
	router.HandleFunc("/api/cryptoPayment", cryptoPaymentHandler).Methods("GET")
	router.HandleFunc("/api/create-checkout-session", createCheckoutSession)
	router.HandleFunc("/api/modify-checkout-session", modifyCheckoutSession)
	router.HandleFunc("/api/session-status", retrieveCheckoutSession)

	http.ListenAndServe(":8090", router)
}
