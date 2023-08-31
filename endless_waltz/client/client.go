package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type Random_Req struct {
	Host string `json:"Host"`
	UUID string `json:"UUID"`
}

func main() {
	log.SetFlags(log.Lshortfile)
	//lets setup our flags here
	msgPtr := flag.String("message", "", "a message to encrypt and send")
	hostPtr := flag.String("host", "localhost", "the server to send the message to")
	randPtr := flag.String("random", "localhost", "the random server to use for pad")
	flag.Parse()

	if len(*msgPtr) > 4096 {
		log.Println("We dont support this yet!")
		return
	}

	conf := &tls.Config{
		// FIx tHis ItS BADDDD
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:6000", *hostPtr), conf)
	if err != nil {
		log.Println(err)
		return
	}

	n, err := conn.Write([]byte("HELO\n"))
	if err != nil {
		log.Println(n, err)
		return
	}

	private_key, err := dh_handshake(conn, "client")
	if err != nil {
		log.Println("Private Key Error!")
		return
	}

	log.Println("Private DH Key: ", private_key)

	//read in response from server
	buf := make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(n, err)
		return
	}
	log.Println(string(buf[:n]))

	//reach out to server and request Pad
	data := Random_Req{
		Host: "client",
		UUID: fmt.Sprintf("%v", string(buf[:n])),
	}
	rapi_data, _ := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	randHost := fmt.Sprintf("http://%s:8090/api/otp", *randPtr)
	req, err := http.NewRequest("POST", randHost, bytes.NewBuffer(rapi_data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		log.Println(error)
		return
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	raw_pad := fmt.Sprintf("%v", res["Pad"])
	cipherText := pad_encrypt(*msgPtr, raw_pad, private_key)
	log.Println(fmt.Sprintf("Ciphertext: %v\n", cipherText))

	n, err = conn.Write([]byte(fmt.Sprintf("%v\n", cipherText)))
	if err != nil {
		log.Println(n, err)
		return
	}

	conn.Close()

}
