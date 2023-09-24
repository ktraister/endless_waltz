package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	ResetColor  = "\033[0m"
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	BlueColor   = "\033[34m"
	PurpleColor = "\033[35m"
	CyanColor   = "\033[36m"
)

type Client_Resp struct {
	UUID string
}

var incomingMsgChan = make(chan string)

// Change handleConnection to act as the "server side" in this transaction
// we'll pass around the websocket to accomplish this
func handleConnection(cm *ConnectionManager, logger *logrus.Logger, configuration Configurations) {
        localUser := fmt.Sprintf("%s_%s", configuration.Server.User, "server")
	_, incoming, err := cm.Read()
	if err != nil {
		logger.Println("Error reading message:", err)
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
		User: configuration.Server.User,
		From: localUser,
		To:   dat["from"].(string),
		Msg:  "HELO",
	}
	b, err := json.Marshal(helo)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cm.Send(b)
	if err != nil {
		logger.Fatal("Server:Unable to write message to websocket: ", err)
		return
	}
	logger.Debug("Responded with HELO")

	user := dat["from"].(string)
	private_key, err := dh_handshake(cm, logger, configuration, "server", user)
	if err != nil {
		logger.Warn("Private Key Error!")
		return
	}
	logger.Debug("Private DH Key: ", private_key)

	//reach out to the api and get our key and pad
	data := []byte(`{"Host": "server"}`)

	//reach out and get random pad and UUID
	req, err := http.NewRequest("POST", configuration.Server.RandomURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.Server.User)
	req.Header.Set("Passwd", configuration.Server.Passwd)
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Error(error)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	pad := fmt.Sprintf("%v", res["Pad"])
	logger.Debug("UUID: ", res["UUID"])

	//send off the UUID to the client
	outgoing := &Message{Type: "UUID",
		User: configuration.Server.User,
		From: configuration.Server.User,
		To:   user,
		Msg:  res["UUID"].(string),
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cm.Send(b)
	if err != nil {
		logger.Fatal("Unable to write message to websocket: ", err)
		return
	}

	logger.Debug("We've just sent off the UUID to client...")

	//receive the encrypted text
	_, incoming, err = cm.Read()
	if err != nil {
		logger.Println("Error reading message:", err)
		return
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return
	}

	logger.Debug("Incoming msg: ", dat["msg"].(string))

	//woof -- this goes away under the messenger paradigm
	// Perform TLS handshake to get the client's certificate
	/*
		if err := tlsConn.Handshake(); err != nil {
			fmt.Println("TLS handshake error:", err)
			return
		}

		//certificate stuff
		clientCert := tlsConn.ConnectionState().PeerCertificates

		var clientCommonName string
		if len(clientCert) == 0 {
			clientCommonName = fmt.Sprintf("%sunknown%s", RedColor, ResetColor)
		} else {
			clientCommonName = fmt.Sprintf("%s%s%s", GreenColor, clientCert[0].Issuer.CommonName, ResetColor)
		}
	*/

	incomingMsgStr := fmt.Sprintf("%s:%s", dat["from"], pad_decrypt(dat["msg"].(string), pad, private_key))
	incomingMsgChan <- incomingMsgStr
}
