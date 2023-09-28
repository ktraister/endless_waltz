package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	User string `json:"user"`
	Msg  string `json:"msg"`
	ok   bool   `json:"ok"`
}

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

func ew_client(logger *logrus.Logger, configuration Configurations, message Post) bool {
	user := fmt.Sprintf("%s_client-%s", configuration.User, uid())
	cm, err := exConnect(logger, configuration, user)
	if err != nil {
		return false
	}
	defer cm.Close()
	passwd := configuration.Passwd
	random := configuration.RandomURL

	targetUser := fmt.Sprintf("%s_%s", string(message.User), "server")

	logger.Debug(fmt.Sprintf("Sending msg %s from user %s to user %s!!", message.Msg, user, targetUser))

	if len(message.Msg) > 4096 {
		logger.Fatal("We dont support this yet!")
		return false
	}

	if passwd == "" || user == "" {
		logger.Fatal("authorized Creds are required")
		return false
	}

	//send HELO to target user
	helo := &Message{Type: "helo",
		User: configuration.User,
		From: user,
		To:   targetUser,
		Msg:  "HELO",
	}
	logger.Debug(helo)
	b, err := json.Marshal(helo)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = cm.Send(b)
	if err != nil {
		logger.Fatal("Client:Unable to write message to websocket: ", err)
		return false
	}
	logger.Debug("Client:Sent init HELO")

	//HELO should be received within 5 seconds to proceed OR exit
	cm.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, incoming, err := cm.Read()
	if err != nil {
		logger.Error("Client:Error reading message:", err)
		return false
	}
	logger.Debug("Client:Read init HELO response")

	err = json.Unmarshal([]byte(incoming), &dat)
	fmt.Println(dat)
	if err != nil {
		logger.Error("Client:Error unmarshalling json:", err)
		return false
	}

	if dat["msg"] == "User not found" {
		logger.Error("Exchange couldn't route a message to ", targetUser)
		return false
	}

	heloUser := strings.Split(dat["from"].(string), "-")[0]
	if dat["msg"] == "HELO" &&
		heloUser == targetUser {
		logger.Debug("Client received HELO from ", heloUser)
	} else {
		logger.Error(fmt.Sprintf("Didn't receive HELO from %s in time, try again later", targetUser))
		return false
	}

	logger.Debug(fmt.Sprintf("Upgrading remote conn user from %s to %s", targetUser, dat["from"].(string)))
	targetUser = dat["from"].(string)

	//reset conn read deadline
	cm.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	//perform DH handshake with the other user
	private_key, err := dh_handshake(cm, logger, configuration, "client", targetUser)
	if err != nil {
		logger.Fatal("Private Key Error!")
		return false
	}

	logger.Info("Private DH Key: ", private_key)

	//read in response from server
	_, incoming, err = cm.Read()
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
	req, err := http.NewRequest("POST", random, bytes.NewBuffer(rapi_data))
	if err != nil {
		logger.Error(err)
		return false
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", passwd)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return false
	}
	json.NewDecoder(resp.Body).Decode(&dat)
	logger.Debug("got response from RandomAPI: ", dat)
	raw_pad := fmt.Sprintf("%v", dat["Pad"])
	cipherText := pad_encrypt(message.Msg, raw_pad, private_key)
	logger.Debug(fmt.Sprintf("Ciphertext: %v\n", cipherText))

	//send the ciphertext to the other user throught the websocket
	outgoing := &Message{Type: "cipher",
		User: configuration.User,
		From: user,
		To:   targetUser,
		Msg:  cipherText,
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = cm.Send(b)
	return true
}
