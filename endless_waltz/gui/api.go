package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func getExUsers(logger *logrus.Logger, configuration Configurations) ([]string, error) {
	urlSlice := strings.Split(configuration.Server.ExchangeURL, "/")
	url := "http://" + urlSlice[2] + "/listUsers"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.Server.User)
	req.Header.Set("Passwd", configuration.Server.Passwd)
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return []string{}, err
	}
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return []string{}, err
	}

	tmpUsers := strings.Split(string(output), ":")
	final := []string{}
	for _, user := range tmpUsers {
		user = strings.Replace(user, " ", "", -1)
		if user != "" && user != configuration.Server.User {
			final = append(final, user)
		}
	}

	return final, nil
}
