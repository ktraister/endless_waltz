package main

import (
    "fmt"
    "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // You can customize the origin check here if needed
        return true
    },
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer conn.Close()

    fmt.Println("Client connected")

    // Handle WebSocket messages
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            fmt.Println(err)
            return
        }

        if messageType == websocket.TextMessage {
            message := string(p)
            fmt.Printf("Received: %s\n", message)

            // Echo the received message back to the client
            err = conn.WriteMessage(websocket.TextMessage, p)
            if err != nil {
                fmt.Println(err)
                return
            }
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleConnection)

    // Start the HTTP server on port 8080 (or any desired port)
    err := http.ListenAndServe(":8081", nil)
    if err != nil {
        fmt.Println(err)
        return
    }
}
