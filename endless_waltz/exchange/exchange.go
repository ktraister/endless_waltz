package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

var clients = make(map[*Client]bool)
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

	ok = checkAuth(req.Header.Get("User"), req.Header.Get("Passwd"), logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	// Create a map to store the slice values
	userMap := make(map[string]struct{})
	for c, _ := range clients {
		user := strings.Split(c.Username, "_")[0]
		// Check if an item is in the slice
		if _, found := userMap[user]; !found {
			logger.Debug("Adding to userlist: ", user)
			userMap[user] = struct{}{}
		}
	}

	userList := ""
	for user, _ := range userMap {
		userList = userList + user + ":"
	}

	logger.Debug(fmt.Sprintf("Returning userlist '%v'", userList))
	w.Write([]byte(userList))
	logger.Info("Someone hit the listUsers route...")
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	user := strings.Split(r.Header.Get("User"), "_")[0]

	ok = checkAuth(user, r.Header.Get("Passwd"), logger)
	if !ok {
		http.Error(w, "403 Unauthorized", http.StatusUnauthorized)
		logger.Info("request denied 403 unauthorized")
		return
	}

	// Ensure client not already connected!
	// Bad things happen :)
	for c, _ := range clients {
		if c.Username == r.Header.Get("User") {
			logger.Warn(fmt.Sprintf("Client %s is already connected, bouncing", c.Username))
			return
		}
	}

	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	// register client
	client := &Client{Conn: ws}
	clients[client] = true
	logger.Info("clients", len(clients), clients, ws.RemoteAddr())

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	receiver(r.Header.Get("User"), client, logger)

	logger.Info("exiting", ws.RemoteAddr().String())
	delete(clients, client)
}

func receiver(user string, client *Client, logger *logrus.Logger) {
	for {
		// read in a message
		// readMessage returns messageType, message, err
		// messageType: 1-> Text Message, 2 -> Binary Message
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
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
		for client := range clients {
			// send message only to involved users
			if client.Username == message.To {
				logger.Info(fmt.Sprintf("Sending message '%s' to client '%s'", message, client.Username))
				err := client.Conn.WriteJSON(message)
				if err != nil {
					logger.Error("Websocket error: ", err)
					client.Conn.Close()
					delete(clients, client)
				}
				sendFlag = 1
			}
			clientStr = clientStr + "," + client.Username
		}
		if sendFlag == 0 {
			logger.Info(fmt.Sprintf("Message '%s' was blackholed because '%s' was not matched in '%s'", message, message.To, clientStr))
			//let the client know here
			for client := range clients {
				// send message only to involved users
				if client.Username == message.From {
					logger.Info(fmt.Sprintf("Sending blackhole message to client '%s'", client.Username))
					message = Message{From: "SYSTEM", To: message.From, Msg: "User not found"}
					err := client.Conn.WriteJSON(message)
					if err != nil {
						logger.Error("Websocket error: ", err)
						client.Conn.Close()
						delete(clients, client)
					}
				}
			}
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
	router.HandleFunc("/listUsers", listUsers)
	router.HandleFunc("/ws", serveWs)
	http.ListenAndServe(":8081", router)
}
