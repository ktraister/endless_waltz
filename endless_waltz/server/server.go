package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Client_Resp struct {
	UUID string
}

func handleConnection(conn net.Conn, logger *logrus.Logger, random_host string) {
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

func main() {
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

	logger := createLogger(configuration.Server.logLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("serverpath is\t\t", configuration.Server.RandomURL)

	cer, err := tls.LoadX509KeyPair(configuration.Server.Cert, configuration.Server.Key)
	if err != nil {
		logger.Fatal(err)
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	//change this to be configurable via config file
	ln, err := tls.Listen("tcp", ":6000", config)
	if err != nil {
		logger.Fatal(err)
		return
	}
	defer ln.Close()

	logger.Info("EW Server is coming online!")
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		go handleConnection(conn, logger, configuration.Server.RandomURL)
	}
}
