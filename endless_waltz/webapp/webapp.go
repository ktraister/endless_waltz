package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func imgHandler(w http.ResponseWriter, req *http.Request) {
	img, err := os.ReadFile(fmt.Sprintf("pages%s", req.URL.Path))
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	img, err := os.ReadFile(fmt.Sprintf("pages%s", req.URL.Path))
	if err != nil {
		return
	}
	w.Write(img)
}


func homePageHandler(w http.ResponseWriter, r *http.Request) {
	// Implement your custom page logic here
	home, err := os.ReadFile("pages/home")
	if err != nil {
		return
	}
	fmt.Fprintln(w, string(home))
}

func signUpPageHandler(w http.ResponseWriter, r *http.Request) {
	signUpForm, err := os.ReadFile("pages/signUp.html")
	if err != nil {
		return
	}

	fmt.Fprintln(w, string(signUpForm))

}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	loginForm, err := os.ReadFile("pages/login.html")
	if err != nil {
		return
	}

	fmt.Fprintln(w, string(loginForm))

}

func signUpHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if req.FormValue("password") != req.FormValue("confirm_password") {
		http.Error(w, "Password did not match confirmation", http.StatusBadRequest)
		return
	}

	//setup database access
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	credential := options.Credential{
		Username: MongoUser,
		Password: MongoPass,
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
	if err != nil {
		logger.Fatal(err)
		return
	} else {
		logger.Info("Database connection succesful!")
	}
	auth_db := client.Database("auth").Collection("keys")

	//check database to ensure username/email ! already exists
	filters := []primitive.M{bson.M{"User": req.FormValue("username")},
		bson.M{"Email": req.FormValue("email")},
	}

	for _,filter := range filters {
	        var result bson.M
		fmt.Println(filter)
		err := auth_db.FindOne(ctx, filter).Decode(&result)
		if err != mongo.ErrNoDocuments {
			//be more specific in future
			http.Error(w, "Username or Email collision", http.StatusBadRequest)
			logger.Error(err)
			return
		}
	}

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	//Write to database with information
	_, err = auth_db.InsertOne(ctx, bson.M{"User":req.FormValue("username"),"Passwd":password,"Email":req.FormValue("email"),"Active":true})
	if err != nil {
		return
	}

	//redirect to main page here pending email confirmation
	http.Redirect(w, req, "/signUpSuccess", http.StatusSeeOther)

}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get username and password from the form
	username := req.FormValue("username")

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	// Authenticate user (you'll need to implement this function)
	if checkAuth(username, password, logger) {
		// Successful login
		// Create a session for the user
		session, _ := store.Get(req, "session-name")
		session.Values["authenticated"] = true
		session.Save(req, w)

		// Redirect to a protected page or display a success message
		http.Redirect(w, req, "/protected", http.StatusSeeOther)
	} else {
		// Failed login
		// Display an error message or redirect to the login page
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}
}

func protectedPageHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	session, err := store.Get(req, "session-name")
	if err != nil {
		logger.Error("sessionNotFound")
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	// Check if the user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		logger.Error("Client unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Display your protected page content here
	fmt.Fprintln(w, "Welcome to the Protected Page!")
}

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("WebApp Server finished starting up!")

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/", homePageHandler).Methods("GET")
	router.HandleFunc("/error", staticHandler).Methods("GET")
	router.HandleFunc("/img/{id}", imgHandler).Methods("GET")
	router.HandleFunc("/login", staticHandler).Methods("GET")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/signUp", staticHandler).Methods("GET")
	router.HandleFunc("/signUp", signUpHandler).Methods("POST")
	router.HandleFunc("/protected", protectedPageHandler).Methods("GET")
	http.ListenAndServe(":8080", router)
}
