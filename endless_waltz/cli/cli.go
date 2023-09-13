package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

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

	fmt.Println(configuration.Server.LogLevel)
	logger := createLogger(configuration.Server.LogLevel, "normal")

	// Reading variables using the model
	logger.Debug("Reading variables using the model..")
	logger.Debug("keypath is\t\t", configuration.Server.Key)
	logger.Debug("crtpath is\t\t", configuration.Server.Cert)
	logger.Debug("serverpath is\t\t", configuration.Server.RandomURL)
	logger.Debug("API_Key is\t\t", configuration.Server.API_Key)

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

	reader := bufio.NewReader(os.Stdin)

	//logger.Info("EW CLI is coming online!")
	for {
		fmt.Print("ew_cli > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" || input == "exit" {
			break
		}

		if input == "help" {
		}

		// old code from the server exe
		/*
			conn, err := ln.Accept()
			if err != nil {
				logger.Error(err)
				continue
			}
			go handleConnection(conn, logger, configuration.Server.RandomURL, configuration.Server.API_Key)
		*/
	}
}
