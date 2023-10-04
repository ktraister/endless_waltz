package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SessionKey")))

type sessionData struct {
	IsAuthenticated bool
	Username        string
}

func imgHandler(w http.ResponseWriter, req *http.Request) {
	img, err := os.ReadFile(fmt.Sprintf("pages%s", req.URL.Path))
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
}

func staticTemplateHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	session, err := store.Get(req, "session-name")

	path := ""
	if req.URL.Path == "/" { 
	    path = "home"
	} else {
	    path = req.URL.Path
        }

	filename := fmt.Sprintf("pages/%s.tmpl", path)

	// Parse the template
	t, err := template.New("").ParseFiles("pages/base.tmpl", filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to parse template")
		return
	}

	var data sessionData
	// Define a data struct for the template
	if session.Values["username"] != nil {
		data = sessionData{
			IsAuthenticated: true,
			Username:        session.Values["username"].(string),
		}
	} else {
		data = sessionData{
			IsAuthenticated: false,
			Username:        "none",
		}
	}

	// Execute the template with the data and write it to the response
	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("Failed to execute template")
		logger.Error(err)
		return
	}

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

func logoutPageHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	session, err := store.Get(req, "session-name")

	//end their session
	session.Options.MaxAge = -1
	err = session.Save(req, w)
	if err != nil {
		logger.Error("Unable to delete session")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}

	logoutForm, err := os.ReadFile("pages/logOutSuccess")
	if err != nil {
		return
	}

	fmt.Fprintln(w, string(logoutForm))

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

	//extensible for other db checks into the future
	//check database to ensure username ! already exists
	filters := []primitive.M{bson.M{"User": req.FormValue("username")}}

	for _, filter := range filters {
		var result bson.M
		fmt.Println(filter)
		err := auth_db.FindOne(ctx, filter).Decode(&result)
		if err != mongo.ErrNoDocuments {
			//be more specific in future
			http.Error(w, "Username collision", http.StatusBadRequest)
			logger.Error(err)
			return
		}
	}

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	//set our signUpTime to the unix time.now()
	now := time.Now()
	signUpTime := fmt.Sprint(now.Unix())

	//Write to database with information
	_, err = auth_db.InsertOne(ctx, bson.M{"User": req.FormValue("username"), "Passwd": password, "SignupTime": signUpTime, "Active": true})
	if err != nil {
		return
	}

	//redirect to main page 5 seconds later using html
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
		session.Values["username"] = username
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
	//fmt.Fprintln(w, fmt.Sprintf("Welcome to the Protected Page, %s!", session.Values["username"]))

	//read in the template
	tmpl, err := os.ReadFile("pages/manageUser.tmpl")
	if err != nil {
		logger.Error("Failed to read template")
		return
	}

	// Parse the template
	t, err := template.New("index").Parse(string(tmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("failed to parse template")
		return
	}

	// Define a data struct for the template
	data := struct {
		IsAuthenticated bool
		Username        string
	}{
		IsAuthenticated: true,
		Username:        session.Values["username"].(string),
	}

	// Execute the template with the data and write it to the response
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("Failed to execute template")
		logger.Error(err)
		return
	}
}

func protectedHandler(w http.ResponseWriter, req *http.Request) {

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

	//delete the user per their request
	ok = deleteUser(logger, session.Values["username"].(string))
	if !ok {
		logger.Error("Unable to delete user")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}

	//end their session
	session.Options.MaxAge = -1
	err = session.Save(req, w)
	if err != nil {
		logger.Error("Unable to delete session")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}

	http.Redirect(w, req, "/deleteSuccess", http.StatusSeeOther)
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
	router.HandleFunc("/", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/error", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/img/{id}", imgHandler).Methods("GET")
	router.HandleFunc("/login", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/signUp", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/signUp", signUpHandler).Methods("POST")
	router.HandleFunc("/signUpSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/deleteSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/protected", protectedPageHandler).Methods("GET")
	router.HandleFunc("/protected", protectedHandler).Methods("POST")
	router.HandleFunc("/downloads", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/how_it_works", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/privacy", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/logout", logoutPageHandler).Methods("GET")
	http.ListenAndServe(":8080", router)
}
