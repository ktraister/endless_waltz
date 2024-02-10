package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/syncmap"
	"net/http"
	"strings"
	"time"
)

var MongoURI string
var MongoUser string
var MongoPass string
var suite = edwards25519.NewBlakeSHA256Ed25519()
var kyberLocalPrivKeys = []kyber.Scalar{}

type rl struct {
	lastRequestTime int64
	requests        int
}

var rateLimitMap = syncmap.Map{}

func rateLimit(user string, limit int) bool {
	value, ok := rateLimitMap.Load(user)
	now := time.Now().Unix()
	//we didn't find the item, no limit
	if !ok {
		rateLimitMap.Store(user, rl{lastRequestTime: now, requests: 1})
		return true
	}

	//typecase here once we're sure we got a struct back
	userRL := value.(rl)

	//no requests yet this second, no limit
	if userRL.lastRequestTime != now {
		rateLimitMap.Store(user, rl{lastRequestTime: now, requests: 1})
		return true
	}

	//request count has reached threshold (5reqs/1sec)
	if userRL.requests == limit {
		return false
	} else {
		//increment and return
		r := userRL.requests + 1
		rateLimitMap.Store(user, rl{lastRequestTime: now, requests: r})
		return true
	}
}

//creating a new qPrivKey in the loop solved a maddening issue
func translatePrivKeys(input string) ([]kyber.Scalar, error) {
	tmp := []kyber.Scalar{}
	for _, v := range strings.Split(input, ",") {
		fmt.Println("Decoding privKey string ", v)
		decodedBytes, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return []kyber.Scalar{}, err
		}

		qPrivKey := suite.Scalar()
		err = qPrivKey.UnmarshalBinary(decodedBytes)
		if err != nil {
			return []kyber.Scalar{}, err
		}
		fmt.Println("Appending Privkey ", qPrivKey)

		tmp = append(tmp, qPrivKey)
		fmt.Println("TMP SLICE ", tmp)
	}

	fmt.Println("Returning privkey map ", tmp)
	return tmp, nil
}

func encryptString(message string, pubKey kyber.Point) (string, error) {
	//encrypt the message
	cipherText, err := ecies.Encrypt(suite, pubKey, []byte(message), suite.Hash)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptString(inString string, privKeyMap []kyber.Scalar) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(inString)
	if err != nil {
		return "", err
	}

	fmt.Sprintf("decoded incoming bytes -->  %d", decodedBytes)

	for _, key := range privKeyMap {
		plainText, err := ecies.Decrypt(suite, key, decodedBytes, suite.Hash)
		if err != nil {
			fmt.Println("Could not decrypt msg with key ", key)
			continue
		} else {
			fmt.Println("decrypted msg with key ", key)
			return string(plainText), nil
		}
	}

	return "", fmt.Errorf("Unable to decrypt message with available pubKeys")
}

// checkAuth needs to get updated to allow users to login with deactive accts
func checkKyberAuth(auth string, logger *logrus.Logger) (string, bool) {
	plainText, err := decryptString(auth, kyberLocalPrivKeys)
	if err != nil {
		logger.Error(err)
		return "", false
	}

	pair := strings.Split(plainText, ":")
	user := pair[0]
	passwd := pair[1]

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
		return "", false
	}

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("checking user '%s' with pass '%s'", user, passwd))
	var filter, result bson.M
	filter = bson.M{"User": user, "Active": true}
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Debug("Found creds in db, checking hash")
	} else if err == mongo.ErrNoDocuments {
		logger.Info("No creds found, unauthorized")
		return "", false
	} else {
		logger.Error(err)
		return "", false
	}

	dbPass := result["Passwd"].(primitive.Binary).Data

	err = bcrypt.CompareHashAndPassword(
		[]byte(dbPass), []byte(passwd))
	return user, err == nil
}

//checkPlainAuth is only used for the web application login -- kyber isnt supported there
func checkPlainAuth(user string, passwd string, active bool, logger *logrus.Logger) bool {
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

	// Defer the close operation to ensure the client is closed when the main function exits
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("error in deferred mongo cleanup func: ", err)
		}
	}()

	auth_db := client.Database("auth").Collection("keys")

	// Check if the item exists in the collection
	logger.Debug(fmt.Sprintf("checking user '%s' with pass '%s'", user, passwd))
	var filter, result bson.M
	if active {
		filter = bson.M{"User": user, "Active": true}
	} else {
		filter = bson.M{"User": user}
	}
	err = auth_db.FindOne(context.TODO(), filter).Decode(&result)
	if err == nil {
		logger.Debug("Found creds in db, checking hash")
	} else if err == mongo.ErrNoDocuments {
		logger.Info("No creds found, unauthorized")
		return false
	} else {
		logger.Error(err)
		return false
	}

	dbPass := result["Passwd"].(primitive.Binary).Data

	err = bcrypt.CompareHashAndPassword(
		[]byte(dbPass), []byte(passwd))
	return err == nil
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
