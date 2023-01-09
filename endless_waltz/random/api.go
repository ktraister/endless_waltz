package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

var jsonMap map[string]interface{}
var dbMap map[string]interface{}
var MongoURI string

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

func base_handler(w http.ResponseWriter, req *http.Request) {
	response := "The base route has been hit successfully!"
	json.NewEncoder(w).Encode(response)
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
			//lets move to using the db to pull an item
			server_resp := Server_Resp{}
			err := otp_db.FindOne(ctx, bson.M{}).Decode(&server_resp)
			if err != nil {
				log.Fatal(err)
			} //else {
			    //lock the item

			fmt.Println(server_resp)

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
			//use above solution to "readOne" of the entries
			UUID := fmt.Sprintf("%v", jsonMap["UUID"])
			filterCursor, err := otp_db.Find(ctx, bson.M{"UUID": UUID})
			if err != nil {
				log.Fatal(err)
			}
			if !filterCursor.Next(ctx) {
				fmt.Println(fmt.Sprintf("No value in Mongo for UUID %v, informing client", jsonMap["UUID"]))
				w.Write([]byte("ERROR: No otp found for UUID included in request."))
				return
			}
			var dbResult []bson.M
			if err = filterCursor.All(ctx, &dbResult); err != nil {
				log.Fatal(err)
			}

			otp := fmt.Sprintf("%v", dbResult[0]["Pad"])
			client_resp := Client_Resp{
				Pad: otp,
			}
			resp, _ := json.Marshal(client_resp)
			if err != nil {
				fmt.Println(err)
			}

			//this is where we respond to the connection
			w.Write(resp)

			//add deletion of mongo pad here
			if _, err = otp_db.DeleteOne(ctx, bson.M{"UUID": UUID}); err != nil {
				fmt.Println(err)
			}
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
	MongoURI = configuration.Server.MongoURI

	router := mux.NewRouter()
	router.HandleFunc("/api/", base_handler).Methods("GET")
	router.HandleFunc("/api/otp", otp_handler).Methods("POST")

	http.ListenAndServe(":8090", router)
}
