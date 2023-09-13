package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Client_Resp struct {
	UUID string
}

func handleConnection(conn net.Conn, logger *logrus.Logger, random_host string, api_key string) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	msg, err := r.ReadString('\n')
	if err != nil {
		println("Uh ohh, error in reading init string...")
		logger.Warn(fmt.Sprintf("Error reading init string: %s", err))
		return
	}

	//new connections should always ask
	if string(msg) != "HELO\n" {
		println("returning...")
		return
	}

	private_key, err := dh_handshake(conn, logger, "server")
	if err != nil {
		logger.Warn("Private Key Error!")
		return
	}
	logger.Debug("Private DH Key: ", private_key)

	//reach out to the api and get our key and pad
	data := []byte(`{"Host": "server"}`)

	//reach out and get random pad and UUID
	req, err := http.NewRequest("POST", random_host, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("API-Key", api_key)
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
	n, err := conn.Write([]byte(fmt.Sprintf("%v", res["UUID"])))
	if err != nil {
		logger.Error(n, err)
		return
	}
	//we should log the client IP at this point
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		logger.Info(addr.IP.String())
	}
	logger.Debug("We've just sent off the UUID to client...")

	//receive the encrypted text
	msg, err = r.ReadString('\n')
	if err != nil {
		logger.Debug("Uh ohh, error in ciphertext string...")
		logger.Warn(err)
		return
	}
	logger.Debug("Incoming msg: ", msg)
	println("decrypted msg")
	println(pad_decrypt(msg, pad, private_key))
}
