package main

import (
	"net/http"
	"os"

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

	http.ListenAndServe(":8090", router)
}
