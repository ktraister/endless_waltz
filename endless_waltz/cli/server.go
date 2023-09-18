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
)

const (
	ResetColor  = "\033[0m"
	RedColor    = "\033[31m"
	GreenColor  = "\033[32m"
	YellowColor = "\033[33m"
	BlueColor   = "\033[34m"
	PurpleColor = "\033[35m"
	CyanColor   = "\033[36m"
)

type Client_Resp struct {
	UUID string
}

func handleConnection(tlsConn *tls.Conn, logger *logrus.Logger, random_host string, api_key string) {
	defer tlsConn.Close()

	r := bufio.NewReader(tlsConn)
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

	private_key, err := dh_handshake(tlsConn, logger, "server")
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
	n, err := tlsConn.Write([]byte(fmt.Sprintf("%v", res["UUID"])))
	if err != nil {
		logger.Error(n, err)
		return
	}
	//we should log the client IP at this point
	addr, ok := tlsConn.RemoteAddr().(*net.TCPAddr)
	if ok {
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

	//woof
	// Perform TLS handshake to get the client's certificate
	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("TLS handshake error:", err)
		return
	}

	//certificate stuff
	clientCert := tlsConn.ConnectionState().PeerCertificates

	var clientCommonName string
	if len(clientCert) == 0 {
		clientCommonName = fmt.Sprintf("%sunknown%s", RedColor, ResetColor)
	} else {
		clientCommonName = fmt.Sprintf("%s%s%s", GreenColor, clientCert[0].Issuer.CommonName, ResetColor)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println(fmt.Sprintf("Receiving msg from %s at host %s...", clientCommonName, addr.IP.String()))
	fmt.Println(pad_decrypt(msg, pad, private_key))
	fmt.Println()
	fmt.Print("EW_cli > ")

}
