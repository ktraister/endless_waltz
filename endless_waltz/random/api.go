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

	ok = checkAuth(req.Header.Get("User"), req.Header.Get("Passwd"), logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	w.Write([]byte("HEALTHY"))
	logger.Info("Someone hit the health check route...")
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
	payload := strings.NewReader(fmt.Sprintf(`{"name":"Endless Waltz Monthly Payment","redirect_url":"%s","pricing_type":"fixed_price","local_price":{"amount":"1.00","currency":"USD"}}`, domain))
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

// below code mostly provided by stripe. Blame them.
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
		//RedirectOnCompletion: stripe.String("never"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				// Provide the exact Price ID (for example, pr_1234) of the product you want to sell
				Price:    stripe.String("price_1OJe1UGcdL8YMSExsJZxn1J1"),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
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

	s, _ := session.Get(req.URL.Query().Get("session_id"), nil)

	writeJSON(w, struct {
		Status        string `json:"status"`
		CustomerEmail string `json:"customer_email"`
	}{
		Status:        string(s.Status),
		CustomerEmail: string(s.CustomerDetails.Email),
	}, logger)
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

//End stripe code

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")
	//stripe.Key = os.Getenv("StripeAPIKey")
	stripe.Key = "sk_test_51O9xNoGcdL8YMSEx9AhtgC768jodZ0DhknQ1KMKLiiXzZQgnxz79ob6JS5qZwrg2cEVVvEimeaXnNMwree7l82hF00zehcsfJc"

	logger := createLogger(LogLevel, LogType)
	logger.Info("Random Server finished starting up!")
	logger.Info(stripe.Key)

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/api/healthcheck", healthHandler).Methods("GET")
	router.HandleFunc("/api/cryptoPayment", cryptoPaymentHandler).Methods("GET")
	router.HandleFunc("/api/create-checkout-session", createCheckoutSession)
	router.HandleFunc("/api/session-status", retrieveCheckoutSession)

	http.ListenAndServe(":8090", router)
}
