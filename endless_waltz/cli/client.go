package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	//"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Random_Req struct {
	Host string `json:"Host"`
	UUID string `json:"UUID"`
}

func ew_client(logger *logrus.Logger, configuration Configurations, message string, host string) {
	//lets setup our flags here
	/*
		msgPtr := flag.String("message", "", "a message to encrypt and send")
		hostPtr := flag.String("host", "localhost", "the server to send the message to")
		randPtr := flag.String("random", "localhost", "the aandom server to use for pad")
		apiKeyPtr := flag.String("API-Key", "", "The API key for the randomAPI server")
		logLvlPtr := flag.String("logLevel", "Warn", "the random server to use for pad")
		flag.Parse()
	*/

	api_key := configuration.Server.API_Key
	random := configuration.Server.RandomURL

	if len(message) > 4096 {
		logger.Fatal("We dont support this yet!")
		return
	}

	if api_key == "" {
		logger.Fatal("Random Servers require an API key")
		return
	}

	//set up certificates
	cert, err := tls.LoadX509KeyPair(configuration.Server.Cert, configuration.Server.Key)
	if err != nil {
		logger.Fatal(err)
	}

	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// FIx tHis ItS BADDDD
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:6000", host), conf)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not connect to host '%s'", host))
		return
	}

	n, err := conn.Write([]byte("HELO\n"))
	if err != nil {
		logger.Fatal(n, err)
		return
	}

	private_key, err := dh_handshake(conn, logger, "client")
	if err != nil {
		logger.Fatal("Private Key Error!")
		return
	}

	logger.Info("Private DH Key: ", private_key)

	//read in response from server
	buf := make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		logger.Fatal(n, err)
		return
	}
	logger.Debug(fmt.Sprintf("got response from server %s", string(buf[:n])))

	//reach out to server and request Pad
	data := Random_Req{
		Host: "client",
		UUID: fmt.Sprintf("%v", string(buf[:n])),
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
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	logger.Debug("got response from RandomAPI: ", res)
	raw_pad := fmt.Sprintf("%v", res["Pad"])
	cipherText := pad_encrypt(message, raw_pad, private_key)
	logger.Debug(fmt.Sprintf("Ciphertext: %v\n", cipherText))

	n, err = conn.Write([]byte(fmt.Sprintf("%v\n", cipherText)))
	if err != nil {
		logger.Fatal(n, err)
		return
	}

	//notify client of successful send
	certs := conn.ConnectionState().PeerCertificates

	var clientCommonName string
	if len(certs) == 0 {
		clientCommonName = fmt.Sprintf("%sunknown%s", RedColor, ResetColor)
	} else {
		clientCommonName = fmt.Sprintf("%s%s%s", GreenColor, certs[0].Issuer.CommonName, ResetColor)
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Sent message successfully to %s at %s", clientCommonName, host))

	conn.Close()

}
