package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"os/signal"
)

var CtlCounter = 0

func trap(conn *websocket.Conn, logger *logrus.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for range c {
		fmt.Println()
		fmt.Println("Ctrl+C Trapped! Use quit to exit or Ctrl+C again.")
		fmt.Println()
		fmt.Print("EW_cli > ")
		CtlCounter++
		if CtlCounter > 1 {
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Fatal(err)
			}
			conn.Close()
			os.Exit(130)
		}
	}

}

func listen(conn *websocket.Conn, logger *logrus.Logger, configuration Configurations) {
	done := make(chan struct{})
	defer close(done)
	for {
		//We need to run our "server" function here
		//server function will need to be able to map incoming message to correct action
		handleConnection(conn, logger, configuration)
	}
}

func main() {
	//configuration stuff
	configuration, err := fetchConfig()
	if err != nil {
		return
	}
	logger := createLogger(configuration.Server.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("randomURL is\t\t", configuration.Server.RandomURL)
	logger.Debug("exchangeURL is\t", configuration.Server.ExchangeURL)
	logger.Debug("user is\t\t", configuration.Server.User)
	logger.Debug("Passwd is\t\t", configuration.Server.Passwd)

	//have the user login every time -- it's no longer APIKeyAuth
	logger.Debug("Checking creds...")
	ok := checkCreds(configuration)
	if !ok {
		return
	}
	logger.Debug("creds passed check!")

	//do some checks and connect to exchange server here
	// Parse the WebSocket URL
	u, err := url.Parse(configuration.Server.ExchangeURL)
	if err != nil {
		logger.Fatal(err)
	}

	// Establish a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"Passwd": []string{configuration.Server.Passwd}, "User": []string{configuration.Server.User}})
	if err != nil {
		logger.Fatal("Could not establish WebSocket connection with ", u.String())
		return
	}
	logger.Debug("Connected to exchange server!")

	defer conn.Close()

	//trap control-c
	go trap(conn, logger)

	//check if user var is empty
	if configuration.Server.User == "" {
		fmt.Println("Can't start without a user...")
		return
	}

	//connect to exchange with our username for mapping
	message := &Message{Type: "startup", User: configuration.Server.User}
	b, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		logger.Fatal(err)
	}

	//this is the interactive part of the EW_cli
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("EW_cli > ")
		raw_input, _ := reader.ReadString('\n')
		input := strings.Split(strings.TrimSpace(raw_input), " ")

		switch input[0] {
		case "":

		case "exit", "quit":
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Fatal(err)
			}
			conn.Close()
			return

		case "listen":
			//listen on conn
			fmt.Println("Listening for incoming messages...")
			for {
				listen(conn, logger, configuration)
			}

		case "help":
			fmt.Println()
			fmt.Println("Help Text")
			fmt.Println("----------------------------")
			fmt.Println()
			fmt.Println("Send and receive messages with other EW_CLIs")
			fmt.Println()

			fmt.Println("exit, quit            ---> leave the CLI")
			fmt.Println("send <user> <message> ---> send a message to an active EW user")
			fmt.Println("help                  ---> print this message")
			fmt.Println()

		case "send":
			if len(input) <= 2 {
				fmt.Println("Not enough fields in send call")
				fmt.Println("Usage: send <user> <message>")
				fmt.Println()
				continue
			}

			msg := ""
			if strings.HasPrefix(input[2], "\"") {
				for i, werd := range input[2:] {
					if i == 0 {
						msg = werd
					} else {
						msg = msg + " " + werd
					}
				}
			} else {
				msg = input[2]
			}

			start := time.Now()
			//this is going to have to change too
			ok := ew_client(logger, configuration, conn, msg, input[1])
			if !ok { fmt.Println("Sending Failed") }
			logger.Info("Sending message duration: ", time.Since(start))
		default:
			fmt.Println("Didn't understand input, try again")
		}

	}

}
