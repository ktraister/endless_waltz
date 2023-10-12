package main

import (
    "regexp"
	"math/rand"
	"time"
	"strings"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWQYZ1234567890"
var disallowed = "{}!@#$%^&*()~`+="

func generateToken() string {
	rand.Seed(time.Now().Unix())
	length := 128
	str := ""

	// Generating Random string
	for i := 0; i < length; i++ {
		str = str + string(charset[rand.Intn(62)])
	}

	// Displaying the random string
	return str
}

func isEmailValid(e string) bool {
    emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
    return emailRegex.MatchString(e)
}

func checkUserInput(input string) bool {
    for i:=0; i < len(input); i++ {
	if strings.Contains(disallowed, string(input[i])) {
           return false
	}
    }
    return true
}
