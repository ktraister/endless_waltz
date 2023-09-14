package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"os/signal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)


func listenForMsg(logger *logrus.Logger, configuration Configurations) {
        //cer, err := tls.LoadX509KeyPair(configuration.Server.Cert, configuration.Server.Key)
        cert, err := tls.LoadX509KeyPair("./certs/example.com.pem", "./certs/example.com.key")

        if err != nil {
                logger.Fatal(err)
                return
        }

        config := &tls.Config{
                Certificates: []tls.Certificate{cert},
                // FIx tHis ItS BADDDD
                InsecureSkipVerify: true,
                //ClientAuth:   tls.RequireAndVerifyClientCert,
                ClientAuth:   tls.RequireAnyClientCert,
        }

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

                // Convert the net.Conn into a TLS connection
                tlsConn, ok := conn.(*tls.Conn)
                if !ok {
                        fmt.Println("Connection is not a TLS connection.")
                        return
                }

                go handleConnection(tlsConn, logger, configuration.Server.RandomURL, configuration.Server.API_Key)
        }
}

func main() {
        //trap control-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
	    for range c {
		fmt.Println()
		fmt.Println("Ctrl+C Trapped! Use quit to exit")
	    }
	}()

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

	//check and make sure inserted API key works
	health_url := fmt.Sprintf("%s%s", strings.Split(configuration.Server.RandomURL, "/otp")[0], "/healthcheck")
	req, err := http.NewRequest("GET", health_url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("API-Key", configuration.Server.API_Key)
	client := &http.Client{}
	resp, error := client.Do(req)
	if error != nil {
		logger.Error(error)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("API Key entered is invalid for randomAPI")
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return
	}

	//goroutine to listen for message
	go listenForMsg(logger, configuration)

	reader := bufio.NewReader(os.Stdin)

	//this is the interactive part of the EW_cli
	for {
		fmt.Print("EW_cli > ")
		raw_input, _ := reader.ReadString('\n')
		input := strings.Split(strings.TrimSpace(raw_input), " ")

		switch input[0] {
		case "":

		case "exit", "quit":
			return

		case "help":
			fmt.Println()
			fmt.Println("Help Text")
			fmt.Println("----------------------------")
			fmt.Println()
			fmt.Println("Send and receive messages with other EW_CLIs")
			fmt.Println()

			fmt.Println("exit, quit            ---> leave the CLI")
			fmt.Println("send <host> <message> ---> send a message to an active EW host")
			fmt.Println("help                  ---> print this message")
			fmt.Println()

		case "send":
			if len(input) <= 2 {
				fmt.Println("Not enough fields in send call")
				fmt.Println("Usage: send <host> <message>")
				fmt.Println()
				continue
			}

			msg := ""
			if strings.HasPrefix(input[2], "\"") {
				for i, werd := range input[2:] {
					if i == 0 {
						msg = werd
					} else {
						msg = msg + " " + werd
					}
				}
			} else {
				msg = input[2]
			}

			ew_client(logger, configuration.Server.API_Key, msg, input[1], configuration.Server.RandomURL)

		default:
			fmt.Println("Didn't understand input, try again")
		}

	}

}
