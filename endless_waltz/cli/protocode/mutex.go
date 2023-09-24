package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnectionManager struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	isClosed bool
}

func (cm *ConnectionManager) Send(message string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isClosed {
		return fmt.Errorf("connection is closed")
	}

	return cm.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func (cm *ConnectionManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isClosed {
		cm.conn.Close()
		cm.isClosed = true
	}
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	connectionManager := &ConnectionManager{
		conn: conn,
	}

	// Example: Send a message
	err = connectionManager.Send("Welcome to the WebSocket server!")
	if err != nil {
		log.Println(err)
		return
	}

	// Example: Receive messages
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Printf("Received message: %s\n", p)

		// Example: Send a response
		response := "Received your message: " + string(p)
		err = connectionManager.Send(response)
		if err != nil {
			log.Println(err)
			break
		}
	}

	connectionManager.Close()
}

func main() {
	http.HandleFunc("/ws", handleWebSocketConnection)
	fmt.Println("WebSocket server is running on :8080/ws")
	http.ListenAndServe(":8080", nil)
}
