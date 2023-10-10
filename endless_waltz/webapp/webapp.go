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

type resetData struct {
	IsAuthenticated bool
	Username        string
	Captcha         bool
	Email           string
	Token           string
}

type sessionData struct {
	IsAuthenticated bool
	Username        string
	Captcha         bool
}

func parseTemplate(logger *logrus.Logger, w http.ResponseWriter, session *sessions.Session, file string) {
	filename := fmt.Sprintf("pages/%s.tmpl", file)

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
			Captcha:         false,
		}
	} else {
		data = sessionData{
			IsAuthenticated: false,
			Username:        "none",
			Captcha:         false,
		}
	}

	//add recaptcha JS to pageif needed
	if file == "/signUp" || file == "/forgotPassword" {
		data.Captcha = true
	}

	// Execute the template with the data and write it to the response
	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("Failed to execute template: ", err)
		return
	}

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
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not get session!")
		return
	}

	path := ""
	if req.URL.Path == "/" {
		path = "home"
	} else {
		path = req.URL.Path
	}

	parseTemplate(logger, w, session, path)
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

	parseTemplate(logger, w, session, "logOutSuccess")
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

	//check recaptcha post here
	logger.Debug(req.FormValue("g-recaptcha-response"))
	ok, err = checkCaptcha(logger, req.FormValue("g-recaptcha-response"))
	if err != nil {
		http.Error(w, "Error performing Recaptcha Check", http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "Recaptcha Check failed", http.StatusBadRequest)
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
		logger.Error("generic mongo signup error: ", err)
		return
	} else {
		logger.Info("Database connection succesful!")
	}
	auth_db := client.Database("auth").Collection("keys")

	//extensible for other db checks into the future
	//check database to ensure username/email ! already exists
	filters := []primitive.M{bson.M{"User": req.FormValue("username")},
		bson.M{"Email": req.FormValue("email")},
	}

	if os.Getenv("ENV") != "local" {
		for _, filter := range filters {
			var result bson.M
			err := auth_db.FindOne(ctx, filter).Decode(&result)
			if err != mongo.ErrNoDocuments {
				//be more specific in future
				http.Error(w, "Collision", http.StatusBadRequest)
				logger.Error("Collision of user/email while checking signup: ", err)
				return
			}
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

	//create a unique token for the user to verify email
	emailVerifyToken := generateToken()

	//send the email before writing to db
	err = sendVerifyEmail(logger, req.FormValue("username"), req.FormValue("email"), emailVerifyToken)
	if err != nil {
		http.Error(w, "Email Verify Fail", http.StatusBadRequest)
		logger.Error("Email verify incoming fail: ", err)
		return
	}

	//Write to database with information
	_, err = auth_db.InsertOne(ctx, bson.M{"User": req.FormValue("username"), "Passwd": password, "SignupTime": signUpTime, "Active": false, "Email": req.FormValue("email"), "EmailVerifyToken": emailVerifyToken})
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		logger.Error("Generic mongo error on user signup write: ", err)
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

func forgotPasswordHandler(w http.ResponseWriter, req *http.Request) {
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

	//check recaptcha post here
	logger.Debug(req.FormValue("g-recaptcha-response"))
	ok, err = checkCaptcha(logger, req.FormValue("g-recaptcha-response"))
	if err != nil {
		http.Error(w, "Error performing Recaptcha Check", http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "Recaptcha Check failed", http.StatusBadRequest)
		return
	}

	//create a unique token for the user to verify email
	emailVerifyToken := generateToken()

	//send the email before writing to db
	err = sendResetEmail(logger, req.FormValue("username"), emailVerifyToken)
	if err == mongo.ErrNoDocuments {
		http.Redirect(w, req, "/forgotPasswordSent", http.StatusSeeOther)
	} else if err != nil {
		http.Error(w, "Email Reset Fail", http.StatusBadRequest)
		logger.Error("Email verify incoming fail: ", err)
		return
	}

	//redirect to main page 5 seconds later
	http.Redirect(w, req, "/forgotPasswordSent", http.StatusSeeOther)

}

//this function should handle a post request with "email" payload
func emailVerifyHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	//we should get form data from our email template... maybe.
	// Parse form data
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if verifyUserSignup(logger, req.FormValue("email"), req.FormValue("user"), req.FormValue("token")) {
		//show the page for user verification success
		http.Redirect(w, req, "/verifySuccess", http.StatusSeeOther)
	} else {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}
}

//this function should handle a post request with "email" payload
func resetPasswordHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	//we should get form data from our email template... maybe.
	// Parse form data
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	if verifyPasswordReset(logger, req.FormValue("email"), req.FormValue("user"), req.FormValue("token")) {
		// Parse the template
		t, err := template.New("").ParseFiles("pages/base.tmpl", "pages/resetPassword.tmpl")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("failed to parse template")
			return
		}

		// Define a data struct for the template
		data := resetData{
			IsAuthenticated: false,
			Username:        req.FormValue("user"),
			Email:           req.FormValue("email"),
			Token:           req.FormValue("token"),
			Captcha:         true,
		}

		// Execute the template with the data and write it to the response
		err = t.ExecuteTemplate(w, "base", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Error("Failed to execute template: ", err)
			return
		}

	} else {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}
}

func resetPasswordSubmitHandler(w http.ResponseWriter, req *http.Request) {
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

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	if submitPasswordReset(logger, req.FormValue("email"), req.FormValue("user"), req.FormValue("token"), password) {
		http.Redirect(w, req, "/resetPasswordSuccess", http.StatusSeeOther)
	} else {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
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

	parseTemplate(logger, w, session, "manageUser")

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
	router.HandleFunc("/verifyEmail", emailVerifyHandler).Methods("POST")
	router.HandleFunc("/verifySuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/signUpSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/deleteSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/protected", protectedPageHandler).Methods("GET")
	router.HandleFunc("/protected", protectedHandler).Methods("POST")
	router.HandleFunc("/downloads", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/how_it_works", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/privacy", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/logout", logoutPageHandler).Methods("GET")
	router.HandleFunc("/forgotPassword", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/forgotPasswordSent", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/resetPassword", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/forgotPassword", forgotPasswordHandler).Methods("POST")
	router.HandleFunc("/resetPassword", resetPasswordHandler).Methods("POST")
	router.HandleFunc("/resetPasswordSubmit", resetPasswordSubmitHandler).Methods("POST")
	router.HandleFunc("/resetPasswordSuccess", staticTemplateHandler).Methods("GET")

	http.ListenAndServe(":8080", router)
}
