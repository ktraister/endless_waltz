package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func handleConnection(conn net.Conn, config *ssh.ServerConfig, logger *logrus.Logger) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		logger.Error("SSH handshake failed: ", err)
		return
	}

	logger.Debug(fmt.Sprintf("User %s authenticated", sshConn.User()))

	// Handle request channels (e.g., session)
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		go handleChannel(newChannel, sshConn.User(), logger)
	}
}

func handleChannel(newChannel ssh.NewChannel, user string, logger *logrus.Logger) {
	if newChannel.ChannelType() != "direct-tcpip" {
		logger.Debug("Unsupported channel type")
		newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		logger.Error("Failed to accept channel: ", err)
		return
	}

	host := "endlesswaltz.xyz"
	port := "443"

	if os.Getenv("ENV") == "local" {
		host = "nginx"
	}

	destConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		logger.Error("Failed to dial destination: ", err)
		newChannel.Reject(ssh.ConnectionFailed, err.Error())
		return
	}

	logger.Info("Proxying HTTPS connection for ", user)

	go func() {
		_, err := io.Copy(destConn, channel)
		if err != nil {
			logger.Error("Error copying data from destination to client:", err)
		}
	}()

	_, err = io.Copy(channel, destConn)
	if err != nil {
		logger.Error("Error copying data from client to destination:", err)
	}

	go func() {
		for req := range requests {
			req.Reply(false, nil)
		}
	}()
}

func main() {
	//setup the logger
	MongoURI = os.Getenv("MongoURI")
	MongoUser = os.Getenv("MongoUser")
	MongoPass = os.Getenv("MongoPass")
	LogLevel := os.Getenv("LogLevel")
	LogType := os.Getenv("LogType")

	logger := createLogger(LogLevel, LogType)

	//specifies global configuration values for SSH algos
	cipherConfig := ssh.Config{
		KeyExchanges: []string{"curve25519-sha256", "curve25519-sha256@libssh.org"},
		Ciphers:      []string{"aes128-gcm@openssh.com", "aes256-gcm@openssh.com", "aes128-ctr", "aes192-ctr", "aes256-ctr"},
		MACs:         []string{"hmac-sha2-256-etm@openssh.com", "hmac-sha2-512-etm@openssh.com", "hmac-sha2-256", "hmac-sha2-512"},
	}

	//SSH server configuration
	sshConfig := &ssh.ServerConfig{
		Config: cipherConfig,
		PasswordCallback: func(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if !rateLimit(c.User(), 1) {
				return nil, fmt.Errorf("RateLimit")
			}
			if checkAuth(c.User(), string(password), true, logger) {
				return nil, nil
			}
			return nil, fmt.Errorf("Password incorrect")
		},
	}

	//Generate a host key - will copy into docker container
	privateBytes, err := os.ReadFile("./keys/private_key")
	if err != nil {
		logger.Error("Error reading Privkey: ", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logger.Error("Error parsing PrivateKey: ", err)
	}
	sshConfig.AddHostKey(private)

	//SSH server listener
	addy := "0.0.0.0:2222"
	listener, err := net.Listen("tcp", addy)
	if err != nil {
		logger.Error("Error creating listener: ", err)
	}
	defer listener.Close()

	logger.Info("Proxy Server startup! Listening on " + addy)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting conn on listener: ", err)
		}

		go handleConnection(conn, sshConfig, logger)
	}
}
