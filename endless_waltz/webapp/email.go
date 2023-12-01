package main

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/smtp"
	"os"
)

type emailData struct {
	FormHost   string
	Username   string
	TargetUser string
	Token      string
}

func templateEmail(logger *logrus.Logger, path string, data emailData) (string, error) {
	filename := fmt.Sprintf("pages/email/%s.tmpl", path)
	logger.Info("templating ", filename)
	// Parse the template
	t, err := template.New("").ParseFiles("pages/email/base.tmpl", filename)
	if err != nil {
		logger.Error("failed to parse email template")
		return "", err
	}

	var rendered bytes.Buffer
	err = t.ExecuteTemplate(&rendered, "base", data)
	if err != nil {
		logger.Error("Error rendering template:", err)
		return "", err
	}

	return rendered.String(), nil
}

func sendSignupEmail(logger *logrus.Logger, username string, targetUser string) error {
	emailUser := os.Getenv("EmailUser")
	emailPass := os.Getenv("EmailPass")

	//connect to our server, set up a message and send it
	auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")

	//stubbed for now, but determine the billing type, and use it to template email
	//checkBillingMethod(logger, username)

	emailContent, err := templateEmail(logger, "signUpTemplate", emailData{})
	if err != nil {
		logger.Error("Unable to template email")
		return err
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Welcome to Endless Waltz\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server")
		return err
	}

	return nil
}

func sendVerifyEmail(logger *logrus.Logger, username string, targetUser string, token string) error {
	emailUser := os.Getenv("EmailUser")
	emailPass := os.Getenv("EmailPass")

	//connect to our server, set up a message and send it
	auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")

	formHost := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		formHost = "https://localhost"
	}

	emailData := emailData{
		FormHost:   formHost,
		Username:   username,
		TargetUser: targetUser,
		Token:      token,
	}

	emailContent, err := templateEmail(logger, "verifyTemplate", emailData)
	if err != nil {
		logger.Error("Unable to template email")
		return err
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Verify Your Email Address\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server")
		return err
	}

	return nil
}

func sendResetEmail(logger *logrus.Logger, username string, token string) error {
	//query the db for the email corresponding to the username
	//and, yknow, make sure the request isn't a bogus user
	targetUser, err := prepareUserPassReset(logger, username, token)
	if err != nil {
		logger.Warn("bogus reset: ", err)
		return err
	}

	emailUser := os.Getenv("EmailUser")
	emailPass := os.Getenv("EmailPass")

	//connect to our server
	auth := smtp.PlainAuth("", emailUser, emailPass, "smtp.gmail.com")

	formHost := "https://endlesswaltz.xyz"
	if os.Getenv("ENV") == "local" {
		formHost = "https://localhost"
	}

	emailData := emailData{
		FormHost:   formHost,
		Username:   username,
		TargetUser: targetUser,
		Token:      token,
	}

	emailContent, err := templateEmail(logger, "resetTemplate", emailData)
	if err != nil {
		logger.Error("Unable to template email")
		return err
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Reset Your Password\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server")
		return err
	}

	return nil
}
