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

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func health_handler(w http.ResponseWriter, req *http.Request) {
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

func crypto_handler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
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

	//update the billing token now that it's been used
	updateFilter := bson.M{"_id": result["_id"]}
	update := bson.M{
		"$set": bson.M{
			"Active":       false,
			"billingToken": generateToken(),
		},
	}
	_, err = db.UpdateOne(ctx, updateFilter, update)
	if err != nil {
		logger.Error("Disable mongo update error: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}

	//create the billing charge
	//https://docs.cloud.coinbase.com/commerce/docs/accepting-crypto#creating-a-charge
	//https://docs.cloud.coinbase.com/commerce/reference/createcharge
	httpClient := &http.Client{}

	payload := strings.NewReader(`{"name":"username"}`)

	cReq, err := http.NewRequest("POST", "https://api.commerce.coinbase.com/charges", payload)
	if err != nil {
		logger.Error("Disable mongo update error: ", err)
		http.Error(w, "500 Error", http.StatusInternalServerError)
		return
	}

	cReq.Header.Add("Content-Type", "application/json")
	cReq.Header.Add("Accept", "application/json")
	cReq.Header.Add("X-CC-Api-Key", os.Getenv("CoinbaseAPIKey"))
	res, err := httpClient.Do(cReq)
	if err != nil {
		logger.Error("Disable mongo update error: ", err)
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

	fmt.Println(string(body))

	//redirect the user

}

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Random Server finished starting up!")

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/api/healthcheck", health_handler).Methods("GET")
	router.HandleFunc("/api/cryptoBilling", crypto_handler).Methods("POST")

	http.ListenAndServe(":8090", router)
}
