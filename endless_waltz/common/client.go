package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"
)

type Message struct {
	Type string `json:"type"`
	User string `json:"user,omitempty"`
	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type Random_Req struct {
	Host string `json:"Host"`
	UUID string `json:"UUID"`
}

var dat map[string]interface{}

func ew_client(logger *logrus.Logger, configuration Configurations, conn *websocket.Conn, message string, targetUser string) bool {
	user := configuration.Server.User
	passwd := configuration.Server.Passwd
	random := configuration.Server.RandomURL

	logger.Debug(fmt.Sprintf("Sending msg %s from user %s to user %s!!", message, user, targetUser))

	if len(message) > 4096 {
		logger.Fatal("We dont support this yet!")
		return false
	}

	if passwd == "" || user == "" {
		logger.Fatal("authorized Creds are required")
		return false
	}

	if targetUser == user {
		fmt.Println("Sending messages to yourself is not allowed")
		return false
	}

	//send HELO to target user
	helo := &Message{Type: "helo",
		User: configuration.Server.User,
		From: configuration.Server.User,
		To:   targetUser,
		Msg:  "HELO",
	}
	logger.Debug(helo)
	b, err := json.Marshal(helo)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		logger.Fatal("Client:Unable to write message to websocket: ", err)
		return false
	}
	logger.Debug("Client:Sent init HELO")

	heloFlag := 0
	//HELO should be received within 5 seconds to proceed OR exit
	//conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, incoming, err := conn.ReadMessage()
	if err != nil {
		logger.Error("Client:Error reading message:", err)
		return false
	}
	logger.Debug("Client:Read init HELO response")

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Client:Error unmarshalling json:", err)
		return false
	}
	
	if dat["msg"] == "User not found" {
		logger.Error("Exchange couldn't route a message to ", targetUser)
		return false
	}


	if dat["msg"] == "HELO" &&
		dat["from"] == targetUser {
		logger.Debug("Client received HELO from ", dat["from"].(string))
		heloFlag = 1
	}

	if heloFlag == 0 {
		logger.Error(fmt.Sprintf("Didn't receive HELO from %s in time, try again later", targetUser))
		return false
	}

	//reset conn read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	//perform DH handshake with the other user
	private_key, err := dh_handshake(conn, logger, configuration, "client", targetUser)
	if err != nil {
		logger.Fatal("Private Key Error!")
		return false
	}

	logger.Info("Private DH Key: ", private_key)

	//read in response from server
	_, incoming, err = conn.ReadMessage()
	if err != nil {
		logger.Error("Error reading message:", err)
		return false
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return false
	}

	logger.Debug(fmt.Sprintf("got response from server %s", dat["msg"]))

	//this will all have to stay the same -- we get the UUID from the "server" above
	//reach out to server and request Pad
	data := Random_Req{
		Host: "client",
		UUID: fmt.Sprintf("%v", dat["msg"]),
	}
	rapi_data, _ := json.Marshal(data)
	if err != nil {
		logger.Warn(err)
	}
	req, err := http.NewRequest("POST", random, bytes.NewBuffer(rapi_data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.Server.User)
	req.Header.Set("Passwd", passwd)
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Fatal(error)
		return false
	}
	json.NewDecoder(resp.Body).Decode(&dat)
	logger.Debug("got response from RandomAPI: ", dat)
	raw_pad := fmt.Sprintf("%v", dat["Pad"])
	cipherText := pad_encrypt(message, raw_pad, private_key)
	logger.Debug(fmt.Sprintf("Ciphertext: %v\n", cipherText))

	//send the ciphertext to the other user throught the websocket
	outgoing := &Message{Type: "cipher",
		User: configuration.Server.User,
		From: configuration.Server.User,
		To:   targetUser,
		Msg:  cipherText,
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = conn.WriteMessage(websocket.TextMessage, b)

	/* Cert stuff needs to change
	certs := conn.ConnectionState().PeerCertificates

	var clientCommonName string
	if len(certs) == 0 {
		clientCommonName = fmt.Sprintf("%sunknown%s", RedColor, ResetColor)
	} else {
		clientCommonName = fmt.Sprintf("%s%s%s", GreenColor, certs[0].Issuer.CommonName, ResetColor)
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Sent message successfully to %s at %s", clientCommonName, host))
	*/

	return true
}

func fetchConfig() (Configurations, error) {
	var configuration Configurations
	//contents of temp config file
	contents := "server:\n  key: \"./certs/server.key\"\n  cert: \"./certs/server.crt\"\n  randomURL: \"http://localhost:8090/api/otp\"\n  exchangeURL: \"ws://localhost:8081/ws\"\n  logLevel: \"Debug\"\n  user: \"KayleighToo\"\n  Passwd: \"arandomnumber\""

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user: ", err)
		return configuration, err
	}

	// Get the user's home directory
	configDir := fmt.Sprintf("%s/.ew", currentUser.HomeDir)
	configFile := fmt.Sprintf("%s/config.yml", configDir)

	//check if directory exists and create if not
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Println("no config dir found, creating...")
		if err := os.Mkdir(configDir, os.ModePerm); err != nil {
			fmt.Println("Unable to create config home dir: ", err)
			return configuration, err
		}
	}

	//check if actual config file exists, create if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("no config file found, creating...")
		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println("Unable to create config file", err)
			return configuration, err
		}

		// Write contents to the file
		_, err = file.WriteString(contents)
		if err != nil {
			fmt.Println("Unable to write temp contents to config file", err)
			return configuration, err
		}
		file.Close()
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.ew/")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Viper:Error reading config file: ", err)
		return configuration, err
	}
	err = viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Println("Viper:Unable to decode into struct: ", err)
		return configuration, err
	}

	return configuration, nil
}

func checkCreds(configuration Configurations) bool {
	//check and make sure inserted creds
	//Random and Exchange will use same mongo, so the creds will be valid for both

	health_url := fmt.Sprintf("%s%s", strings.Split(configuration.Server.RandomURL, "/otp")[0], "/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.Server.User)
	req.Header.Set("Passwd", configuration.Server.Passwd)
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Could not connect to configured randomAPI ", configuration.Server.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false
	}
	if resp == nil {
		fmt.Println("Could not connect to configured randomAPI ", configuration.Server.RandomURL)
		fmt.Println("Quietly exiting now. Please reconfigure.")
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("creds entered are invalid for randomAPI")
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return false
	}
	return true
}
