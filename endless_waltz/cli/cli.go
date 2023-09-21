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
	"github.com/spf13/viper"
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
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	var configuration Configurations
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	logger := createLogger(configuration.Server.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("randomURL is\t\t", configuration.Server.RandomURL)
	logger.Debug("exchangeURL is\t", configuration.Server.ExchangeURL)
	logger.Debug("user is\t\t", configuration.Server.User)
	logger.Debug("API_Key is\t\t", configuration.Server.API_Key)

	//check and make sure inserted API key works
	//Random and Exchange will use same mongo, so the API key will be valid for both
	logger.Debug("Checking api key...")
	health_url := fmt.Sprintf("%s%s", strings.Split(configuration.Server.RandomURL, "/otp")[0], "/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("API-Key", configuration.Server.API_Key)
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not connect to configured randomAPI ", configuration.Server.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return
	}
	if resp == nil {
		fmt.Println("Could not connect to configured randomAPI ", configuration.Server.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("API Key entered is invalid for randomAPI")
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return
	}
	logger.Debug("API Key passed check!")

	//do some checks and connect to exchange server here
	// Parse the WebSocket URL
	u, err := url.Parse(configuration.Server.ExchangeURL)
	if err != nil {
		logger.Fatal(err)
	}

	// Establish a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Could not establish WebSocket connection with ", u.String())
		return
	}
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
			listen(conn, logger, configuration)

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
			ew_client(logger, configuration, conn, msg, input[1])
			logger.Info("Sending message duration: ", time.Since(start))
			ew_client(logger, configuration, conn, msg, input[1])

		default:
			fmt.Println("Didn't understand input, try again")
		}

	}

}
