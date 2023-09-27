package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Client_Resp struct {
	UUID string
}

var incomingMsgChan = make(chan Post)

// Change handleConnection to act as the "server side" in this transaction
// we'll pass around the websocket to accomplish this
func handleConnection(cm *ConnectionManager, logger *logrus.Logger, configuration Configurations) {
	localUser := fmt.Sprintf("%s_%s", configuration.User, "server")
	_, incoming, err := cm.Read()
	if err != nil {
		logger.Error("Error reading message:", err)
		return
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return
	}

	//new connections should always ask
	if dat["msg"] == "HELO" {
		logger.Debug("Received HELO from ", dat["from"])
	} else {
		logger.Warn("New connection didn't HELO, bouncing")
		return
	}

	//we need to respond with a HELO here
	helo := &Message{Type: "helo",
		User: configuration.User,
		From: localUser,
		To:   dat["from"].(string),
		Msg:  "HELO",
	}
	b, err := json.Marshal(helo)
	if err != nil {
		logger.Error(err)
		return
	}

	err = cm.Send(b)
	if err != nil {
		logger.Error("Server:Unable to write message to websocket: ", err)
		return
	}
	logger.Debug("Responded with HELO")

	user := dat["from"].(string)
	private_key, err := dh_handshake(cm, logger, configuration, "server", user)
	if err != nil {
		logger.Error("Private Key Error!")
		return
	}
	logger.Debug("Private DH Key: ", private_key)

	//reach out to the api and get our key and pad
	data := []byte(`{"Host": "server"}`)

	//reach out and get random pad and UUID
	req, err := http.NewRequest("POST", configuration.RandomURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", configuration.Passwd)
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Error(error)
		return
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	pad := fmt.Sprintf("%v", res["Pad"])
	logger.Debug("UUID: ", res["UUID"])

	//send off the UUID to the client
	outgoing := &Message{Type: "UUID",
		User: configuration.User,
		From: configuration.User,
		To:   user,
		Msg:  res["UUID"].(string),
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		logger.Error(err)
		return
	}

	err = cm.Send(b)
	if err != nil {
		logger.Error("Unable to write message to websocket: ", err)
		return
	}

	logger.Debug("We've just sent off the UUID to client...")

	//receive the encrypted text
	_, incoming, err = cm.Read()
	if err != nil {
		logger.Error("Error reading message:", err)
		return
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return
	}

	logger.Debug("Incoming msg: ", dat["msg"].(string))

	incomingMsg := Post{User: dat["user"].(string), Msg: pad_decrypt(dat["msg"].(string), pad, private_key), ok: true}
	incomingMsgChan <- incomingMsg
}
