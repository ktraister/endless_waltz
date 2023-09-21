package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

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
var MongoURI string
var MongoUser string
var MongoPass string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

func serveWs(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("logger").(*logrus.Logger)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error("Could not configure logger!")
		return
	}

	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
	}

	client := &Client{Conn: ws}
	// register client
	clients[client] = true
	logger.Debug("clients", len(clients), clients, ws.RemoteAddr())

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	receiver(client, logger)

	logger.Info("exiting", ws.RemoteAddr().String())
	delete(clients, client)
}

func receiver(client *Client, logger *logrus.Logger) {
	for {
		// read in a message
		// readMessage returns messageType, message, err
		// messageType: 1-> Text Message, 2 -> Binary Message
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			logger.Error(err)
			return
		}

		m := &Message{}

		err = json.Unmarshal(p, m)
		if err != nil {
			logger.Error("error while unmarshaling chat", err)
			continue
		}

		logger.Info("host", client.Conn.RemoteAddr())
		if m.Type == "startup" {
			// do mapping on startup
			client.Username = m.User
			logger.Info("client successfully mapped", &client, client, client.Username)
		} else {
			logger.Info("received message", m.Type, m.Msg)
			//broadcast <- &c
			broadcast <- *m

		}
	}
}

func broadcaster(logger *logrus.Logger) {
	for {
		message := <-broadcast
		for client := range clients {
			// send message only to involved users
			if client.Username == message.To {
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
	router.HandleFunc("/ws", serveWs)
	http.ListenAndServe(":8081", router)
}