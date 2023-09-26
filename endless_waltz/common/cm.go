package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
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

func exConnect(logger *logrus.Logger, configuration Configurations, ctype string) (*ConnectionManager, error) {
	user := configuration.Server.User + "_" + ctype

	// Parse the WebSocket URL
	u, err := url.Parse(configuration.Server.ExchangeURL)
	if err != nil {
		logger.Fatal(err)
	}

	// Establish a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Passwd": []string{configuration.Server.Passwd}, "User": []string{user}})
	if err != nil {
		logger.Fatal("Could not establish WebSocket connection with ", u.String())
		return &ConnectionManager{}, err
	}
	logger.Debug("Connected to exchange server!")

	connectionManager := &ConnectionManager{
		conn: conn,
	}

	//connect to exchange with our username for mapping
	message := &Message{Type: "startup", User: user}
	b, err := json.Marshal(message)
	if err != nil {
		logger.Fatal(err)
		return &ConnectionManager{}, err
	}
	err = connectionManager.Send(b)
	if err != nil {
		logger.Fatal(err)
		return &ConnectionManager{}, err
	}

	return connectionManager, nil
}
