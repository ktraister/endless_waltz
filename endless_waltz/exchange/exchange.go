package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/syncmap"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	Conn     *websocket.Conn
	Username string
}

type Message struct {
	Type string `json:"type"`
	User string `json:"user,omitempty"`
	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

var clients = syncmap.Map{}
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func listUsers(w http.ResponseWriter, req *http.Request) {
	logger, ok := req.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	ok = rateLimit(req.Header.Get("User"), 2)
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

	// Create a map to store the slice values
	users := []string{}
	clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		user := strings.Split(client.Username, "_")[0]
		// Check if an item is in the slice
		if !slices.Contains(users, user) {
			logger.Debug("Adding to userlist: ", user)
			users = append(users, user)
		}

		// this will continue iterating
		return true
	})

	userList := ""
	for _, user := range users {
		userList = userList + user + ":"
	}

	logger.Debug(fmt.Sprintf("Returning userlist '%v'", userList))
	w.Write([]byte(userList))
	logger.Debug("Someone hit the listUsers route...")
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	user := strings.Split(r.Header.Get("User"), "_")[0]

	ok = rateLimit(user, 3)
	if !ok {
		http.Error(w, "429 Rate Limit", http.StatusTooManyRequests)
		logger.Info("request denied 429 rate limit")
		return
	}

	ok = checkAuth(user, r.Header.Get("Passwd"), logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	// Ensure client not already connected!
	// Bad things happen :)
	bounceFlag := false
	clients.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		if client.Username == r.Header.Get("User") {
		        logger.Warn(fmt.Sprintf("Client %s is already connected, bouncing", client.Username))
			bounceFlag = true
			return false
		}
		return true
	})
	if bounceFlag {
		return
	}

	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	// register client
	client := &Client{Conn: ws}
	clients.Store(client, true)
	logger.Info("registered new client; ", client, ws.RemoteAddr())

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	receiver(r.Header.Get("User"), client, logger)

	logger.Debug("exiting", ws.RemoteAddr().String())
	clients.Delete(client)
}

func receiver(user string, client *Client, logger *logrus.Logger) {
	for {
		// read in a message
		// readMessage returns messageType, message, err
		// messageType: 1-> Text Message, 2 -> Binary Message
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				client.Conn.Close()
				break
			}
			logger.Error(err)
			continue
		}

		m := &Message{}

		err = json.Unmarshal(p, m)
		if err != nil {
			logger.Error("error while unmarshaling chat", err)
			continue
		}

		if m.Type == "startup" {
			// do mapping on startup
			client.Username = m.User
			logger.Info("client successfully mapped", &client, client)
		} else {
			logger.Debug("received message, broadcasting: ", m)
			broadcast <- *m
		}
	}
}

func broadcaster(logger *logrus.Logger) {
	for {
		message := <-broadcast
		//don't use my relays to send shit to yourself
		if message.To == message.From {
			logger.Warn(fmt.Sprintf("Possible abuse from %s: refusing to self send message on relay", message.To))
			continue
		}
		sendFlag := 0
		clientStr := ""
		clients.Range(func(key, value interface{}) bool {
			client := key.(*Client)
			// send message only to involved users
			if client.Username == message.To {
				logger.Debug(fmt.Sprintf("Sending message '%s' to client '%s'", message, client.Username))
				err := client.Conn.WriteJSON(message)
				if err != nil {
					logger.Error("Websocket error: ", err)
					client.Conn.Close()
					clients.Delete(key)
				}
				sendFlag = 1
				return false
			}
			clientStr = clientStr + "," + client.Username
			return true
		})
		if sendFlag == 0 {
			logger.Info(fmt.Sprintf("Message '%s' was blackholed because '%s' was not matched in '%s'", message, message.To, clientStr))
			//let the client know here
			clients.Range(func(key, value interface{}) bool {
				client := key.(*Client)
				// send message only to involved users
				if client.Username == message.From {
					logger.Debug(fmt.Sprintf("Sending blackhole message to client '%s'", client.Username))
					message = Message{From: "SYSTEM", To: message.From, Msg: "User not found"}
					err := client.Conn.WriteJSON(message)
					if err != nil {
						logger.Error("Websocket error: ", err)
						client.Conn.Close()
						clients.Delete(key)
					}
				}
				return true
			})
		}
	}
}

func main() {
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)
	logger.Info("Exchange Server finished starting up!")

	go broadcaster(logger)

	router := mux.NewRouter()
	router.Use(LoggerMiddleware(logger))
	router.HandleFunc("/ws/listUsers", listUsers).Methods("GET")
	router.HandleFunc("/ws", serveWs)
	http.ListenAndServe(":8081", router)
}
