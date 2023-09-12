package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type Random_Req struct {
	Host string `json:"Host"`
	UUID string `json:"UUID"`
}

func main() {
	//lets setup our flags here
	msgPtr := flag.String("message", "", "a message to encrypt and send")
	hostPtr := flag.String("host", "localhost", "the server to send the message to")
	randPtr := flag.String("random", "localhost", "the random server to use for pad")
	logLvlPtr := flag.String("logLevel", "Warn", "the random server to use for pad")
	flag.Parse()

	logger := createLogger(*logLvlPtr, "normal")

	fmt.Println(fmt.Sprintf("Sending message to %s...", *hostPtr))

	if len(*msgPtr) > 4096 {
		logger.Fatal("We dont support this yet!")
		return
	}

	conf := &tls.Config{
		// FIx tHis ItS BADDDD
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:6000", *hostPtr), conf)
	if err != nil {
		logger.Fatal("Could not connect to remote message server...")
		return
	}

	//check here if the random API is up and ready
	randHost := fmt.Sprintf("http://%s:8090/api/healthcheck", *randPtr)
	req, err := http.NewRequest("GET", randHost)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Fatal(error)
		return
	}

	if resp.Body == "HEALTHY" {
	    logger.Info("Random host healthy, Proceeding")
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

	logger.Debug("Private DH Key: ", private_key)

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
	randHost = fmt.Sprintf("http://%s:8090/api/otp", *randPtr)
	req, err = http.NewRequest("POST", randHost, bytes.NewBuffer(rapi_data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client = &http.Client{}
	resp, error = client.Do(req)
	if error != nil {
		logger.Fatal(error)
		return
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	raw_pad := fmt.Sprintf("%v", res["Pad"])
	cipherText := pad_encrypt(*msgPtr, raw_pad, private_key)
	logger.Debug(fmt.Sprintf("Ciphertext: %v\n", cipherText))

	n, err = conn.Write([]byte(fmt.Sprintf("%v\n", cipherText)))
	if err != nil {
		logger.Fatal(n, err)
		return
	}

	//notify client of successful send
	fmt.Println("Sent message successfully!")
	fmt.Println("goodbye :)")
	//logger

	conn.Close()

}
