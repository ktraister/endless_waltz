package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	serverURL = "ws://localhost:8080/ws" // Change this to your WebSocket server URL
)

func main() {
	// Create a WebSocket dialer
	dialer := websocket.DefaultDialer

	// Connect to the WebSocket server
	conn, _, err := dialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create channels for communication
	readChannel := make(chan string)
	writeChannel := make(chan string)

	// Handle incoming messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				close(readChannel)
				return
			}
			readChannel <- string(message)
		}
	}()

	// Handle outgoing messages
	go func() {
		for {
			select {
			case message, ok := <-writeChannel:
				if !ok {
					return
				}
				err := conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					log.Println("Error sending message:", err)
					return
				}
			}
		}
	}()

	// Create a channel for capturing signals (e.g., Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-interrupt:
			fmt.Println("Interrupt received, closing WebSocket connection.")
			close(writeChannel)
			return
		case message := <-readChannel:
			fmt.Printf("Received message: %s\n", message)
		case <-time.After(5 * time.Second):
			// Example: Send a message every 5 seconds
			message := "Hello, server!"
			writeChannel <- message
		}
	}
}
