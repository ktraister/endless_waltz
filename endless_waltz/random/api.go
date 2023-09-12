package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var MongoURI string
var MongoUser string
var MongoPass string

var jsonMap map[string]interface{} // XXX is this data in a well defined JSON structure? does it change often?

type Server_Resp struct {
	UUID string
	Pad  string
}

type Client_Resp struct {
	Pad string
}

type Error_Resp struct {
	Error string
}

func checkAPIKey(remote_key string, logger *logrus.Logger) bool {
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
		logger.Fatal(err)
		return false
	}
	auth_db := client.Database("auth").Collection("keys")

	// Define a filter to match all documents (an empty filter)
	filter := bson.M{}

	// Create a cursor to retrieve documents
	cursor, err := auth_db.Find(context.Background(), filter)
	if err != nil {
		logger.Fatal(err)
	}

	defer cursor.Close(context.Background())

	// Loop through the cursor to read documents
	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			logger.Fatal(err)
		}

		//this will print every item returned 
		//this is where we can match&return true
		//remote_key
		logger.Info(result) // Print the document
		return true
	}

	if err := cursor.Err(); err != nil {
		logger.Fatal(err)
	}

	// allow all to pass for now
	return true
}

// Custom middleware function to inject a logger into the request context
func LoggerMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Inject the logger into the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "logger", logger)
			r = r.WithContext(ctx)

			// Call the next middleware or handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}

func health_handler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("ERROR: Could not configure logger!")
		return
	}
	w.Write([]byte("HEALTHY"))
	logger.Info("Someone hit the health check route...")

}

func otp_handler(w http.ResponseWriter, req *http.Request) {
	//logging setup
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("ERROR: Could not configure logger!")
		return
	}

	reqBody, err := io.ReadAll(req.Body) // newer versions of go moved ReadAll to io instead of ioutil
	if err != nil {
		logger.Error(err)
	}

	//logging our header will show IP once server is in AWS
	logger.Info(fmt.Sprintf("Incoming request: %s, %s\n", req.Header.Get("X-Forwarded-For"), reqBody))

	if len(reqBody) == 0 {
		logger.Debug("Found no body for this request, returning")
		//lets return a different error code here -- not sure what
		w.WriteHeader(404)
	} else {
		logger.Debug("Found body for the request, proceeding!\n")
		json.Unmarshal([]byte(reqBody), &jsonMap)

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
			logger.Fatal(err)
			return
		}
		otp_db := client.Database("otp").Collection("otp")

		host := jsonMap["Host"]
		uuid := jsonMap["UUID"]
		switch {
		case host == "server":
			//lets move to using the db to pull an item
			server_resp := Server_Resp{}
			err := otp_db.FindOne(ctx, bson.M{"LOCK": nil}).Decode(&server_resp)
			if err != nil {
				logger.Error(err)
			} else {
				//lock the item
				uuid, _ := primitive.ObjectIDFromHex(server_resp.UUID)
				filter := bson.D{{"UUID", uuid}}
				update := bson.D{{"$set", bson.D{{"LOCK", "true"}}}}
				_, err := otp_db.UpdateOne(ctx, filter, update)
				if err != nil {
					logger.Error(err)
					return
				}
			}

			resp, _ := json.Marshal(server_resp)
			if err != nil {
				logger.Error(err)
				return
			}

			//this is where we respond to the connection
			w.Write(resp)
		case host == "client" && uuid == nil:
			logger.Warn(fmt.Sprintf("No UUID value in request, informing client"))
			w.Write([]byte("ERROR: No UUID included in request."))
			return
		case host == "client":
			//mongo
			//https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
			//use above solution to "readOne" of the entries
			UUID := fmt.Sprintf("%v", jsonMap["UUID"])
			filterCursor, err := otp_db.Find(ctx, bson.M{"UUID": UUID})
			if err != nil {
				logger.Error(err)
				return
			}
			if !filterCursor.Next(ctx) {
				logger.Warn(fmt.Sprintf("No value in Mongo for UUID %v, informing client", jsonMap["UUID"]))
				w.Write([]byte("ERROR: No otp found for UUID included in request."))
				return
			}
			var dbResult []bson.M
			if err = filterCursor.All(ctx, &dbResult); err != nil {
				logger.Error(err)
			}

			otp := fmt.Sprintf("%v", dbResult[0]["Pad"])
			client_resp := Client_Resp{
				Pad: otp,
			}
			resp, _ := json.Marshal(client_resp)
			if err != nil {
				logger.Warn(err)
				return
			}

			//this is where we respond to the connection
			w.Write(resp)

			//add deletion of mongo pad here
			if _, err = otp_db.DeleteOne(ctx, bson.M{"UUID": UUID}); err != nil {
				logger.Error(err)
				return
			}
		}
	}
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
	router.HandleFunc("/api/healthcheck", health_handler).Methods("GET") // XXX is this intended to behave like /ping would? like an et phone home?
	router.HandleFunc("/api/otp", otp_handler).Methods("POST")

	http.ListenAndServe(":8090", router)
}
