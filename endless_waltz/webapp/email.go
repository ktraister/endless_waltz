package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
)

func sendVerifyEmail(logger *logrus.Logger, username string, targetUser string, token string) error {
	emailUser := os.Getenv("EmailUser")
	emailPass := os.Getenv("EmailPass")

	//connect to our server, set up a message and send it
	auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")

	//grab our template
	fileContent, err := ioutil.ReadFile("./pages/email/template")
	if err != nil {
		log.Fatal("Unable to read email template")
		return err
	}

	formHost := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		formHost = "http://localhost:8080"
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Welcome to Endless Waltz\r\n" +
		mime +
		fmt.Sprintf(string(fileContent), formHost, username, targetUser, token))

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		log.Fatal("unable to send email to gmail server")
		return err
	}

	return nil
}
