package main

import (
	"log"
        "os"
	"net/smtp"
)

func sendVerifyEmail(logger *logrus.Logger, targetUser string, token string) error {
        emailUser := os.Getenv("emailUser")
        emailPass := os.Getenv("emailPass")

	//connect to our server, set up a message and send it
	auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")

	from := emailUser
        to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n" +
		"Subject: Why aren’t you using Endless Waltz yet?\r\n" +
		"\r\n" +
		"Here’s the space for our token:\r\n %s \r\n", targetUser, token)

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
