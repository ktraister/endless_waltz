package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
)

var jsonMap map[string]interface{} // XXX is this data in a well defined JSON structure? does it change often?
var MongoURI string
var MongoUser string
var MongoPass string

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
	w.Write([]byte("The base route has been hit successfully!"))
}

func otp_handler(w http.ResponseWriter, req *http.Request) {
	reqBody, err := io.ReadAll(req.Body) // newer versions of go moved ReadAll to io instead of ioutil
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
		credential := options.Credential{
			Username: MongoUser,
			Password: MongoPass,
		}
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI).SetAuth(credential))
		if err != nil {
			log.Println(err)
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
				log.Println(err)
			} else {
				//lock the item
				uuid, _ := primitive.ObjectIDFromHex(server_resp.UUID)
				filter := bson.D{{"UUID", uuid}}
				update := bson.D{{"$set", bson.D{{"LOCK", "true"}}}}
				result, err := otp_db.UpdateOne(ctx, filter, update)
				if err != nil {
					log.Println(err)
				} else {
					log.Println(result)
				}
			}

			log.Println(server_resp)

			resp, _ := json.Marshal(server_resp)
			if err != nil {
				log.Println(err)
				return
			}

			//this is where we respond to the connection
			w.Write(resp)
		case host == "client" && uuid == nil:
			log.Println(fmt.Sprintf("No UUID value in request, informing client"))
			w.Write([]byte("ERROR: No UUID included in request."))
			return
		case host == "client":
			//mongo
			//https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
			//use above solution to "readOne" of the entries
			UUID := fmt.Sprintf("%v", jsonMap["UUID"])
			filterCursor, err := otp_db.Find(ctx, bson.M{"UUID": UUID})
			if err != nil {
				log.Fatal(err)
				return
			}
			if !filterCursor.Next(ctx) {
				log.Println(fmt.Sprintf("No value in Mongo for UUID %v, informing client", jsonMap["UUID"]))
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
				log.Println(err)
				return
			}

			//this is where we respond to the connection
			w.Write(resp)

			//add deletion of mongo pad here
			if _, err = otp_db.DeleteOne(ctx, bson.M{"UUID": UUID}); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func main() {

	log.Println("Random server coming online!")
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")

	router := mux.NewRouter()
	router.HandleFunc("/api/", base_handler).Methods("GET") // XXX is this intended to behave like /ping would? like an et phone home?
	router.HandleFunc("/api/otp", otp_handler).Methods("POST")

	http.ListenAndServe(":8090", router)
}
