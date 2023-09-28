package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
	"sort"
)

func removeIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func getExUsers(logger *logrus.Logger, configuration Configurations) ([]string, error) {
	urlSlice := strings.Split(configuration.ExchangeURL, "/")
	url := "http://" + urlSlice[2] + "/listUsers"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User", configuration.User)
	req.Header.Set("Passwd", configuration.Passwd)
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
		if user != "" && user != configuration.User {
			final = append(final, user)
		}
	}

	//sort the list in alphabetical order
	sort.Strings(final)

	return final, nil
}
