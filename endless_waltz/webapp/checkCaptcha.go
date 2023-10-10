package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type ResponseStruct struct {
	Success  bool     `json:"success"`
	TS       string   `json:"challenge_ts"`
	Hostname string   `json:"hostname"`
	EC       []string `json:"error-codes"`
}

func checkCaptcha(logger *logrus.Logger, input string) (bool, error) {
	requestURL := "https://www.google.com/recaptcha/api/siteverify"
	payload := fmt.Sprintf("secret=%s&response=%s", os.Getenv("CaptchaKey"), input)

	// Make the POST request
	response, err := http.Post(requestURL, "application/x-www-form-urlencoded; charset=utf-8", bytes.NewBuffer([]byte(payload)))
	if err != nil {

		return false, err
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP request failed with status code: %d", response.StatusCode)
	}

	// Decode the JSON response
	var responseBody ResponseStruct
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return false, err
	}

	if responseBody.Success == false {
		logger.Warn("User failed captcha on signup: ", responseBody.EC)
		return false, nil
	} else {
		return true, nil
	}
}
