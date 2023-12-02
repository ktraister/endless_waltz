package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v76"
	"html/template"
	"net/http"
	"os"
	"strings"
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
	TemplateTag     string
}

type sessionData struct {
	IsAuthenticated bool
	Username        string
	Captcha         bool
	Stripe          bool
	CaptchaFail     bool
	TemplateTag     string
	Email           string
}

func parseTemplate(logger *logrus.Logger, w http.ResponseWriter, req *http.Request, session *sessions.Session, file string) {
	filename := fmt.Sprintf("pages/%s.tmpl", file)

	// Parse the template
	t, err := template.New("").ParseFiles("pages/base.tmpl", filename)
	if err != nil {
		logger.Error("failed to parse template: ", err)
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//ensure values are set
	if session.Values["username"] == nil {
		session.Values["username"] = ""
	}
	if session.Values["authenticated"] == nil {
		session.Values["authenticated"] = false
	}
	if session.Values["email"] == nil {
		session.Values["email"] = ""
	}

	session.Save(req, w)

	var data sessionData
	// Define a data struct for the template
	data = sessionData{
		IsAuthenticated: session.Values["authenticated"].(bool),
		Username:        session.Values["username"].(string),
		Captcha:         false,
		Stripe:          false,
		TemplateTag:     csrf.Token(req),
		Email:           session.Values["email"].(string),
	}

	//add recaptcha JS to pageif needed
	if file == "/signUp" || file == "/forgotPassword" {
		data.Captcha = true
	}

	//add stripe js where needed
	if file == "/signUp" {
		data.Stripe = true
	}

	//add capcha fail if needed
	if session.Values["CaptchaFail"] == true {
		data.CaptchaFail = true
	}

	// Execute the template with the data and write it to the response
	err = t.ExecuteTemplate(w, "base", data)
	if err != nil {
		logger.Error("Failed to execute template: ", err)
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

}

func imgHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in imgHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	path := fmt.Sprintf("pages%s", req.URL.Path)

	_, err := os.Stat(path)
	//return !os.IsNotExist(err)
	if errors.Is(err, os.ErrNotExist) {
		return
	}

	img, err := os.ReadFile(path)
	if err != nil {
		logger.Error("Failed to serve Img: ", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(img)
}

func staticTemplateHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in staticTemplateHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	session, _ := store.Get(req, "session-name")

	path := ""
	if req.URL.Path == "/" {
		path = "home"
	} else {
		path = req.URL.Path
	}

	parseTemplate(logger, w, req, session, path)
}

func logoutPageHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in logoutPageHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	session, err := store.Get(req, "session-name")
	if err != nil {
		logger.Error("Could not get session in logoutPageHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//end their session
	session.Options.MaxAge = -1
	session.Values["authenticated"] = false
	err = session.Save(req, w)
	if err != nil {
		logger.Error("Unable to delete session")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}

	parseTemplate(logger, w, req, session, "logOutSuccess")
}

func signUpHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in signUpHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	session, err := store.Get(req, "session-name")
	if err != nil {
		logger.Error("Could not get session in logoutPageHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err = req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in signUpHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//confirm passwords match
	if req.FormValue("password") != req.FormValue("confirm_password") {
		http.Redirect(w, req, "/signUp", http.StatusSeeOther)
		return
	}

	if !isPasswordValid(req.FormValue("password")) {
		http.Redirect(w, req, "/signUp", http.StatusSeeOther)
		return
	}

	//check for special characters in username
	username := strings.ToLower(req.FormValue("username"))
	ok = checkUserInput(username)
	if !ok {
		http.Redirect(w, req, "/signUp", http.StatusSeeOther)
		return
	}

	//check email is valid
	ok = isEmailValid(req.FormValue("email"))
	if !ok {
		http.Redirect(w, req, "/signUp", http.StatusSeeOther)
		return
	}

	//check recaptcha post here
	logger.Debug(req.FormValue("g-recaptcha-response"))
	ok, err = checkCaptcha(logger, req.FormValue("g-recaptcha-response"))
	if err != nil {
		logger.Error("Error while checking captcha response: ", err)
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}
	if !ok {
		logger.Warn("Captcha check invalid for: ", username)
		session.Values["CaptchaFail"] = true
		session.Save(req, w)
		logger.Debug("CaptchaFail, redirecting: ", session.Values["CaptchaFail"])
		http.Redirect(w, req, "/signUp", http.StatusFound)
		return
	}

	//reset session values just in case
	session.Values["CaptchaFail"] = false
	session.Save(req, w)

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
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	} else {
		logger.Info("Database connection succesful!")
	}
	auth_db := client.Database("auth").Collection("keys")

	//extensible for other db checks into the future
	//check database to ensure username/email ! already exists
	//removed email check for now. Have as many accts as you want
	filters := []primitive.M{bson.M{"User": username}, bson.M{"Email": req.FormValue("email")}}

	//run our extensible DB checks
	for i, filter := range filters {
		var result bson.M
		err := auth_db.FindOne(ctx, filter).Decode(&result)
		if err != mongo.ErrNoDocuments {
			switch i {
			case 0:
				session.Values["username"] = username
				logger.Debug("username in use: ", username)
			case 1:
				session.Values["email"] = req.FormValue("email")
				logger.Debug("email in use: ", req.FormValue("email"))
			}
			session.Save(req, w)
			http.Redirect(w, req, "/signUp", http.StatusFound)
			return
		}
	}

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	//set our signUpTime to the unix time.now()
	signUpTime := fmt.Sprint(time.Now().Unix())

	//create a unique token for the user to verify email
	emailVerifyToken := generateToken()

	//send the email before writing to db
	err = sendVerifyEmail(logger, username, req.FormValue("email"), emailVerifyToken)
	if err != nil {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		logger.Error("Email verify outgoing fail: ", err)
		return
	}

	today := time.Now()
	threshold := today.Add(168 * time.Hour).Format("01-02-2006")
	if req.FormValue("payment") == "crypto" {
		//Write to database with information
		//we'll need to update this insert depending on card/crypto payment
		_, err = auth_db.InsertOne(ctx, bson.M{"User": username,
			"Passwd":              password,
			"SignupTime":          signUpTime,
			"Active":              false,
			"Email":               req.FormValue("email"),
			"EmailVerifyToken":    emailVerifyToken,
			"cryptoBilling":       true,
			"billingCycleEnd":     threshold,
			"billingEmailSent":    false,
			"billingCyclePaid":    false,
			"billingReminderSent": false,
			"billingToken":        generateToken(),
		})
		if err != nil {
			http.Redirect(w, req, "/error", http.StatusSeeOther)
			logger.Error("Generic mongo error on user signup write: ", err)
			return
		}
	} else {
		logger.Error("Unrecognized Input")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//redirect to main page 5 seconds later using html
	http.Redirect(w, req, "/signUpSuccess", http.StatusSeeOther)

}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in loginHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in loginHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Get username and password from the form
	username := strings.ToLower(req.FormValue("username"))

	//check for special characters in username
	ok = checkUserInput(req.FormValue("username"))
	if !ok {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	ok = rateLimit(req.FormValue("username"), 1)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	//create our hasher to hash our pass
	hash := sha512.New()
	hash.Write([]byte(req.FormValue("password")))
	hashSum := hash.Sum(nil)
	password := hex.EncodeToString(hashSum)

	//api_lib checkAuth function
	if checkAuth(username, password, logger) {
		//create a session for the user
		session, _ := store.Get(req, "session-name")
		session.Values["authenticated"] = true
		session.Values["username"] = username
		session.Save(req, w)

		//redirect to protected page
		http.Redirect(w, req, "/protected", http.StatusSeeOther)
	} else {
		//redirect to the login page
		http.Redirect(w, req, "/login", http.StatusSeeOther)
	}
}

func forgotPasswordHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in forgotPasswordHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in forgotPasswordHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//check for special characters in username
	username := strings.ToLower(req.FormValue("username"))
	ok = checkUserInput(username)
	if !ok {
		http.Redirect(w, req, "/forgotPassword", http.StatusSeeOther)
		return
	}

	//check recaptcha post here
	logger.Debug(req.FormValue("g-recaptcha-response"))
	ok, err = checkCaptcha(logger, req.FormValue("g-recaptcha-response"))
	if err != nil {
		logger.Error("Error while checking captcha response: ", err)
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}
	if !ok {
		logger.Warn("Recaptcha Check failed for user: ", username)
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//create a unique token for the user to verify email
	emailVerifyToken := generateToken()

	//send the email before writing to db
	err = sendResetEmail(logger, username, emailVerifyToken)
	if err == mongo.ErrNoDocuments {
		//if no documents returned, a reset has been requested for
		//non-existent user. Return the same link as normal.
		http.Redirect(w, req, "/forgotPasswordSent", http.StatusSeeOther)
		return
	} else if err != nil {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		logger.Error("forgotPasswordHandler fail: ", err)
		return
	}

	//redirect to main page 5 seconds later
	http.Redirect(w, req, "/forgotPasswordSent", http.StatusSeeOther)

}

// this function should handle a post request with "email" payload
func emailVerifyHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in emailVerifyHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	var email, user, token string
	// Parse form data
	for k, v := range req.URL.Query() {
		switch k {
		case "email":
			email = v[0]
		case "user":
			user = v[0]
		case "token":
			token = v[0]
		}
	}

	//check for special characters in username
	ok = checkUserInput(user)
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//check email is valid
	ok = isEmailValid(email)
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	if verifyUserSignup(logger, email, user, token) {
		//show the page for user verification success
		http.Redirect(w, req, "/verifySuccess", http.StatusSeeOther)
	} else {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}
}

// this function should handle a post request with "email" payload
func resetPasswordHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in resetPasswordHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//need to do static template if no/incorrect form data
	// Parse form data
	var email, user, token string
	for k, v := range req.URL.Query() {
		switch k {
		case "email":
			email = v[0]
		case "user":
			user = v[0]
		case "token":
			token = v[0]
		}
	}

	//check for special characters in username
	ok = checkUserInput(user)
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//check email is valid
	ok = isEmailValid(email)
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	if verifyPasswordReset(logger, email, user, token) {
		// Parse the template
		t, err := template.New("").ParseFiles("pages/base.tmpl", "pages/resetPassword.tmpl")
		if err != nil {
			logger.Error("failed to parse template in passwordResetHandler")
			http.Redirect(w, req, "/error", http.StatusSeeOther)
			return
		}

		// Define a data struct for the template
		data := resetData{
			IsAuthenticated: false,
			Username:        user,
			Email:           email,
			Token:           token,
			Captcha:         true,
			TemplateTag:     csrf.Token(req),
		}

		// Execute the template with the data and write it to the response
		err = t.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Error("Failed to execute template: ", err)
			http.Redirect(w, req, "/error", http.StatusSeeOther)
			return
		}

	} else {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
	}
}

func resetPasswordSubmitHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in resetPasswordSubmitHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in resetPasswordSubmitHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//check for special characters in username
	ok = checkUserInput(req.FormValue("user"))
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	//check email is valid
	ok = isEmailValid(req.FormValue("email"))
	if !ok {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	if !isPasswordValid(req.FormValue("password")) {
		http.Redirect(w, req, "/error", http.StatusSeeOther)
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
		logger.Error("Could not configure logger in protectedPageHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in protectedPageHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	session, err := store.Get(req, "session-name")
	if err != nil {
		logger.Error("Could not get session in protectedPageHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Check if the user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		logger.Error("Client unauthorized")
		http.Redirect(w, req, "/unauthorized", http.StatusSeeOther)
		return
	}

	parseTemplate(logger, w, req, session, "manageUser")

}

func protectedHandler(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		logger.Error("Could not configure logger in protectedHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := req.ParseForm()
	if err != nil {
		logger.Error("Failed to parse form data in protectedHandler")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	session, err := store.Get(req, "session-name")
	if err != nil {
		logger.Error("Could not get session in protectedHandler!")
		http.Redirect(w, req, "/error", http.StatusSeeOther)
		return
	}

	// Check if the user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, req, "/unauthorized", http.StatusSeeOther)
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

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	logger, _ := req.Context().Value("logger").(*logrus.Logger)
	session, _ := store.Get(req, "session-name")
	parseTemplate(logger, w, req, session, "not_found")
}

func main() {
	CSRFAuthKey := []byte(os.Getenv("CSRFAuthKey"))
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")
	stripe.Key = "sk_test_51O9xNoGcdL8YMSEx9AhtgC768jodZ0DhknQ1KMKLiiXzZQgnxz79ob6JS5qZwrg2cEVVvEimeaXnNMwree7l82hF00zehcsfJc"
	//stripe.Key := os.Getenv("StripeAPIKey")

	logger := createLogger(LogLevel, LogType)
	logger.Info("WebApp Server finished starting up!")

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.HandleFunc("/", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/error", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/img/{id}", imgHandler).Methods("GET")
	router.HandleFunc("/login", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/signUp", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/signUp", signUpHandler).Methods("POST")
	router.HandleFunc("/verifyEmail", emailVerifyHandler).Methods("GET")
	router.HandleFunc("/verifySuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/signUpSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/deleteSuccess", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/protected", protectedPageHandler).Methods("GET")
	router.HandleFunc("/protected", protectedHandler).Methods("POST")
	router.HandleFunc("/downloads", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/unauthorized", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/how_it_works", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/privacy", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/privacy_policy", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/terms_and_conditions", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/logout", logoutPageHandler).Methods("GET")
	router.HandleFunc("/forgotPassword", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/forgotPassword", forgotPasswordHandler).Methods("POST")
	router.HandleFunc("/forgotPasswordSent", staticTemplateHandler).Methods("GET")
	router.HandleFunc("/resetPassword", resetPasswordHandler).Methods("GET")
	router.HandleFunc("/resetPasswordSubmit", resetPasswordSubmitHandler).Methods("POST")
	router.HandleFunc("/resetPasswordSuccess", staticTemplateHandler).Methods("GET")

	http.ListenAndServe(":8080", csrf.Protect(CSRFAuthKey)(router))
}
