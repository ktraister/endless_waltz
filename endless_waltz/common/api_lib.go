package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var MongoURI string
var MongoUser string
var MongoPass string

func checkAuth(user string, passwd string, logger *logrus.Logger) bool {
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
		logger.Error(err)
		return false
	}

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("checking user '%s' with pass '%s'", user, passwd))
	filter := bson.M{"Passwd": passwd, "User": user, "Active": true}
	var result bson.M
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Info("Found creds in db, authorized")
		return true
	} else if err == mongo.ErrNoDocuments {
		logger.Warn("No creds found, unauthorized")
		return false
	} else {
		logger.Error(err)
		return false
	}
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
