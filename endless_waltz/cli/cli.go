package main

import (
	"crypto/tls"
	"fmt"
	"bufio"
	"strings"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func listenForMsg(logger *logrus.Logger, configuration Configurations) {
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
		go handleConnection(conn, logger, configuration.Server.RandomURL, configuration.Server.API_Key)
	}
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

	logger := createLogger(configuration.Server.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("serverpath is\t\t", configuration.Server.RandomURL)
	logger.Debug("API_Key is\t\t", configuration.Server.API_Key)

	//goroutine to listen for message
	go listenForMsg(logger, configuration)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("EW_cli > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}
	}

}
