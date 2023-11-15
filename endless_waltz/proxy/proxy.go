package main

import (
    "fmt"
    "os"
    "io"
    "net"
    "net/http"
    "strings"
    "crypto/sha512"
    "encoding/hex" 

    "golang.org/x/crypto/ssh"
    "github.com/sirupsen/logrus"
)

func hashPass(password []byte) string {
        //create our hasher to hash our pass
        hash := sha512.New()    
        hash.Write(password)
        hashSum := hash.Sum(nil)
        hashString := hex.EncodeToString(hashSum)
        return hashString
}

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
        go handleChannel(newChannel, logger)
    }
}

func splitHostPort(data string) (string, string) {
    data = strings.Split(data, "r")[0]
    data = strings.Split(data, "\\")[0]
    return data, "80"
}

func handleChannel(newChannel ssh.NewChannel, logger *logrus.Logger) {
    logger.Error(newChannel.ChannelType())
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

    host := "icanhazip.com"
    port := "80"
    logger.Debug(fmt.Sprintf("net.Split --> host: %s port: %s", host, port))
    if err != nil {
	    logger.Error("Failed to split host and port: ", err)
	    newChannel.Reject(ssh.ConnectionFailed, err.Error())
	    return
    }

    req := &http.Request{
	    Method: "CONNECT",
	    Host:   net.JoinHostPort(host, port),
    }       

    destConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
    if err != nil {
	    logger.Error("Failed to dial destination: %v", err)
	    newChannel.Reject(ssh.ConnectionFailed, err.Error())
	    return
    }
    
    logger.Debug("Proxying HTTPS connection to %s", req.Host)
    

    go func() {
	    defer channel.Close()
	    defer destConn.Close()

	    go io.Copy(destConn, channel)
	    io.Copy(channel, destConn)
    }()

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

    //SSH server configuration
    sshConfig := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
            if checkAuth(c.User(), hashPass(password), logger) {
                return nil, nil
            }
            return nil, fmt.Errorf("Password incorrect")
        },
    }

    //Generate a host key - will copy into docker container
    privateBytes, err := os.ReadFile("private_key")
    if err != nil {
        logger.Error(err)
    }
    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        logger.Error(err)
    }
    sshConfig.AddHostKey(private)

    //SSH server listener
    addy := "0.0.0.0:2222"
    listener, err := net.Listen("tcp", addy)
    if err != nil {
        logger.Error(err)
    }
    defer listener.Close()

    logger.Info("Proxy Server startup! Listening on " + addy)

    for {
        conn, err := listener.Accept()
        if err != nil {
            logger.Error(err)
        }

        go handleConnection(conn, sshConfig, logger)
    }
}
