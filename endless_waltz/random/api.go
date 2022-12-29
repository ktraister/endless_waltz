package main

import (
    "fmt"
    "log"
    "context"
    "time"
    "encoding/json"
    "net/http"
    "io/ioutil"
    "math/rand"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/gorilla/mux"

    "github.com/google/uuid"
    "github.com/spf13/viper"
)

var jsonMap map[string]interface{}
var dbMap map[string]interface{}
var MongoURI string
var UploadAPIKey string
const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

func random_pad() string{
    b := make([]byte, 500)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func base_handler(w http.ResponseWriter, req *http.Request) {
    response := "The base route has been hit successfully!"
    json.NewEncoder(w).Encode(response)
}

func upload_handler(w http.ResponseWriter, req *http.Request) {
    reqBody, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Fatal(err)
    }

    //check the key and pass if we don't match
    json.Unmarshal([]byte(reqBody), &jsonMap) 
    if jsonMap["APIKey"] != UploadAPIKey {
        fmt.Printf("Incoming bad request: %s, %s\n", req.Header.Get("X-Forwarded-For"), reqBody)
	w.Write([]byte("API Key did not match >:("))
	return
    }

    //logging our header will show IP once server is in AWS
    fmt.Printf("Incoming good request: %s, %s\n", req.Header.Get("X-Forwarded-For"), reqBody)

    //connect to mongo
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
    if err != nil {
	fmt.Println(err)
    }
    otp_db := client.Database("otp").Collection("otp")

    //check and see how many entries are in DB. 
    // if too many entries, return "pass"

    //temp code
    pad := "foo"

    //uuid create and send to db
    id := uuid.New().String()
    _, err = otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"otp", pad}}) 
    if err != nil {
	fmt.Println(err)
    }

    w.Write([]byte("success"))
}

func otp_handler(w http.ResponseWriter, req *http.Request) {
    reqBody, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Fatal(err)
    }

    //logging our header will show IP once server is in AWS
    fmt.Printf("Incoming request: %s, %s\n", req.Header.Get("X-Forwarded-For"), reqBody)
    fmt.Printf("%s\n", reqBody)    

    if len(reqBody) == 0 {
	fmt.Printf("Found no body for this request, returning")
	w.WriteHeader(404)
    } else {
	fmt.Printf("Found body for the request, proceeding!\n")
        json.Unmarshal([]byte(reqBody), &jsonMap) 
	
	//connect to mongo
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
	    fmt.Println(err)
	}
	otp_db := client.Database("otp").Collection("otp")

	if jsonMap["Host"] == "server" {
	    //this should happen inside the api to prevent servers from setting UUIDs for attacks
	    //generate uuids for server and write to redis
	    id := uuid.New().String()
	    pad := random_pad()

	    //here's where we write the pad and UUID to mongo
	    _, err := otp_db.InsertOne(ctx, bson.D{{"UUID", id}, {"otp", pad}}) 
	    if err != nil {
		fmt.Println(err)
	    }

	    server_resp := Server_Resp {
		UUID: id,
		Pad: pad,
            }
	    resp, _ := json.Marshal(server_resp)
	    if err != nil {
		fmt.Println(err)
	    }

	    //this is where we respond to the connection
	    w.Write(resp)
	    
        } else if jsonMap["Host"] == "client" {
	    if jsonMap["UUID"] == nil {
		fmt.Println(fmt.Sprintf("No UUID value in request, informing client"))
		w.Write([]byte("ERROR: No UUID included in request."))
		return
            }

	    //mongo
	    //https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
	    UUID := fmt.Sprintf("%v", jsonMap["UUID"])
	    filterCursor, err := otp_db.Find(ctx, bson.M{"UUID": UUID})
	    if err != nil {
		log.Fatal(err)
	    }
	    if ! filterCursor.Next(ctx) { 
                fmt.Println(fmt.Sprintf("No value in Mongo for UUID %v, informing client", jsonMap["UUID"]))
		w.Write([]byte("ERROR: No otp found for UUID included in request."))
		return
	    } 
	    var dbResult []bson.M
	    if err = filterCursor.All(ctx, &dbResult); err != nil {
		log.Fatal(err)
	    }

	    //add deletion of mongo pad here

	    otp := fmt.Sprintf("%v", dbResult[0]["otp"])
	    client_resp := Client_Resp {
		Pad: otp, 
            }
	    resp, _ := json.Marshal(client_resp)
	    if err != nil {
		fmt.Println(err)
	    }

	    //this is where we respond to the connection
	    w.Write(resp)

        }
    }
}

func main() {

    fmt.Println("Random server coming online!")
    //configuration stuff
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.SetConfigType("yml")
    var configuration Configurations
    if err := viper.ReadInConfig(); err != nil {
        fmt.Printf("Error reading config file, %s", err)
    }
    err := viper.Unmarshal(&configuration)
    if err != nil {
            fmt.Printf("Unable to decode into struct, %v", err)
    }

    // Reading variables using the model
    fmt.Println("Reading variables using the model..")
    fmt.Println("MongoURI is\t\t", configuration.Server.MongoURI)
    fmt.Println("UploadAPIKey is\t\t", configuration.Server.UploadAPIKey)
    MongoURI = configuration.Server.MongoURI
    UploadAPIKey = configuration.Server.UploadAPIKey

    router := mux.NewRouter()
    router.HandleFunc("/api/", base_handler).Methods("GET")
    router.HandleFunc("/api/otp", otp_handler).Methods("POST")
    router.HandleFunc("/api/uploads", upload_handler).Methods("POST")

    http.ListenAndServe(":8090", router)
}
