package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/spf13/viper"
)

type Client_Resp struct {
	UUID string
}

func handleConnection(conn net.Conn, random_host string) {
	r := bufio.NewReader(conn)
	msg, err := r.ReadString('\n')
	if err != nil {
		println("Uh ohh, error in reading init string...")
		log.Println(err)
		return
	}

	//new connections should always ask
	if string(msg) != "HELO\n" {
		println("returning...")
		return
	}

	private_key := dh_handshake(conn, "server") 
	if private_key == 0 {
	    fmt.Println("Private Key Error!")
	    return
	} 

	//reach out to the api and get our key and pad
	data := []byte(`{"Host": "server"}`)

	//reach out and get random pad and UUID
	req, err := http.NewRequest("POST", random_host, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		panic(error)
	}
	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	pad := fmt.Sprintf("%v", res["Pad"])
	fmt.Println("Pad: %v", res["Pad"])
	fmt.Println("UUID: %v", res["UUID"])

	//send off the UUID to the client
	n, err := conn.Write([]byte(fmt.Sprintf("%v", res["UUID"])))
	if err != nil {
		log.Println(n, err)
		return
	}
	//we should log the client IP at this point
	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		fmt.Println(addr.IP.String())
	}
	println("We've just sent off the UUID to client...")

	//receive the encrypted text
	msg, err = r.ReadString('\n')
	if err != nil {
		println("Uh ohh, error in ciphertext string...")
		log.Println(err)
		return
	}
	fmt.Println("Incoming msg: ", msg)
	println("decrypted msg")
	println(pad_decrypt(msg, pad))

	conn.Close()
}

func main() {
	log.SetFlags(log.Lshortfile)

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

	// Reading variables using the model
	fmt.Println("Reading variables using the model..")
	fmt.Println("keypath is\t\t", configuration.Server.Key)
	fmt.Println("crtpath is\t\t", configuration.Server.Cert)
	fmt.Println("serverpath is\t\t", configuration.Server.RandomURL)

	cer, err := tls.LoadX509KeyPair(configuration.Server.Cert, configuration.Server.Key)
	if err != nil {
		log.Println(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":6000", config)
	if err != nil {
		log.Println(err)
		return
	}
	defer ln.Close()

	fmt.Println("EW Server is coming online!")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, configuration.Server.RandomURL)
	}
}
