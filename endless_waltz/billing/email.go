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
	filename := fmt.Sprintf("email/%s.tmpl", path)
	logger.Info("templating ", filename)
	// Parse the template
	t, err := template.New("").ParseFiles("email/base.tmpl", filename)
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

func sendCryptoBillingEmail(logger *logrus.Logger, username string, targetUser string, token string) {
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

	emailContent, err := templateEmail(logger, "billingTemplate", emailData)
	if err != nil {
		logger.Error("Unable to template email: ", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Monthly Billing\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server: ", err)
	}
}

func sendCryptoBillingReminder(logger *logrus.Logger, username string, targetUser string, token string) {
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

	emailContent, err := templateEmail(logger, "billingReminderTemplate", emailData)
	if err != nil {
		logger.Error("Unable to template email: ", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Monthly Billing Reminder\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server: ", err)
	}
}

func sendCryptoBillingDisabled(logger *logrus.Logger, username string, targetUser string, token string) {
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

	emailContent, err := templateEmail(logger, "disableTemplate", emailData)
	if err != nil {
		logger.Error("Unable to template email: ", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	from := emailUser
	to := []string{targetUser}
	msg := []byte(fmt.Sprintf("To: %s\r\n", targetUser) +
		"Subject: Your Account Has Been Disabled\r\n" +
		mime +
		emailContent)

	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)

	if err != nil {
		logger.Error("unable to send email to gmail server: ", err)
	}

}

//send crypto payment thank you email
