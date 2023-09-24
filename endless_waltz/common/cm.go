package main

import (
    "sync"
    "fmt"
    "github.com/gorilla/websocket"
)

//start CM
type ConnectionManager struct {
        conn     *websocket.Conn
        mu       sync.Mutex
        isClosed bool
}
 
func (cm *ConnectionManager) Send(message []byte) error {
        cm.mu.Lock()

        if cm.isClosed {
                return fmt.Errorf("connection is closed")
        }   

	err := cm.conn.WriteMessage(websocket.TextMessage, []byte(message))
        cm.mu.Unlock()

	return err
}           

func (cm *ConnectionManager) Read() (int, []byte, error) {
        cm.mu.Lock()
            
        if cm.isClosed {
                return 0, []byte{}, fmt.Errorf("connection is closed")
        }   

	i, b, err := cm.conn.ReadMessage()
        cm.mu.Unlock()

	return i, b, err
}           
            
func (cm *ConnectionManager) Close() {
        cm.mu.Lock()
        defer cm.mu.Unlock()
            
        if !cm.isClosed {
                cm.conn.Close()
                cm.isClosed = true
        }   
} 

