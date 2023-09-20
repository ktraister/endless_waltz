package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	//"time"
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

func ew_client(logger *logrus.Logger, configuration Configurations, conn *websocket.Conn, message string, user string) {
	api_key := configuration.Server.API_Key
	random := configuration.Server.RandomURL

	if len(message) > 4096 {
		logger.Fatal("We dont support this yet!")
		return
	}

	if api_key == "" {
		logger.Fatal("authorized API keys are required")
		return
	}

	//send HELO to target user
	//n, err := conn.Write([]byte("HELO\n"))
	helo := &Message{Type: "helo",
		User: configuration.Server.User,
		From: configuration.Server.User,
		To:   user,
		Msg:  "HELO",
	}
	b, err := json.Marshal(helo)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		logger.Fatal("Client:Unable to write message to websocket: ", err)
		return
	}
	logger.Debug("Client:Sent init HELO")

	//commenting out the HELO loop -- seems to be causing problems
	/*
		heloFlag := 0
		//HELO should be received within 5 seconds to proceed OR exit
		for start := time.Now(); time.Since(start) < time.Second*5; {
			_, incoming, err := conn.ReadMessage()
			if err != nil {
			    logger.Error("Client:Error reading message:", err)
				return
			}

			err = json.Unmarshal([]byte(incoming), &dat)
			if err != nil {
			    logger.Error("Client:Error unmarshalling json:", err)
				return
			}

			if dat["msg"] == "HELO" &&
				dat["from"] == user {
				logger.Debug("Client received HELO from ", dat["from"].(string))
				heloFlag = 1
			} else {
				break
			}
		}

		if heloFlag == 0 {
			logger.Error(fmt.Sprintf("Didn't receive HELO from %s in time, try again later", user))
			return
		}
	*/

	_, incoming, err := conn.ReadMessage()
	if err != nil {
		logger.Error("Client:Error reading message:", err)
		return
	}
	logger.Debug(incoming)

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Client:Error unmarshalling json:", err)
		return
	}

	if dat["msg"] == "HELO" &&
		dat["from"] == user {
		logger.Debug("Client received HELO from ", dat["from"].(string))
	} else {
		logger.Error(fmt.Sprintf("Didn't receive HELO from %s in time, try again later", user))
		return

	}

	//perform DH handshake with the other user
	private_key, err := dh_handshake(conn, logger, configuration, "client", user)
	if err != nil {
		logger.Fatal("Private Key Error!")
		return
	}

	logger.Info("Private DH Key: ", private_key)

	//read in response from server
	_, incoming, err = conn.ReadMessage()
	if err != nil {
		logger.Error("Error reading message:", err)
		return
	}

	err = json.Unmarshal([]byte(incoming), &dat)
	if err != nil {
		logger.Error("Error unmarshalling json:", err)
		return
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
	req.Header.Set("API-Key", api_key)
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Fatal(error)
		return
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
		To:   user,
		Msg:  cipherText,
	}
	b, err = json.Marshal(outgoing)
	if err != nil {
		fmt.Println(err)
		return
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

	conn.Close()

}
