package main

import (
    "fmt"
    "github.com/gorilla/websocket"
    "log"
    "net/url"
    "os"
    "os/signal"
)

func main() {
    // Define the WebSocket server URL
    serverURL := "ws://localhost:8081/ws" // Replace with your WebSocket server URL

    // Parse the WebSocket URL
    u, err := url.Parse(serverURL)
    if err != nil {
        log.Fatal(err)
    }

    // Establish a WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create a channel to handle incoming messages
    done := make(chan struct{})
    go func() {
        defer close(done)
        for {
            _, message, err := conn.ReadMessage()
            if err != nil {
                log.Println("Error reading message:", err)
                return
            }
            fmt.Printf("Received: %s\n", message)
        }
    }()

    // Send a message to the server (optional)
    message := []byte("{\"Type\":\"bootup\", \"user\":\"foo\"}")

    err = conn.WriteMessage(websocket.TextMessage, message)
    if err != nil {
        log.Fatal(err)
    }

    message = []byte("{\"type\":\"test\", \"chat\":{\"id\":\"123456\",\"from\":\"foo\",\"to\":\"Kayleigh\",\"message\":\"sending init message\"}}")

    err = conn.WriteMessage(websocket.TextMessage, message)
    if err != nil {
        log.Fatal(err)
    }

    // Handle signals to gracefully close the connection
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    select {
    case <-interrupt:
        fmt.Println("Interrupt received. Closing WebSocket connection...")
        err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
        if err != nil {
            log.Fatal(err)
        }
        conn.Close()
    case <-done:
    }
}
