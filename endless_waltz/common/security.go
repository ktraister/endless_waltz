package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWQYZ1234567890"
var disallowed = "{}!@#$%^&*()~`+="
var day, month, year int

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

func isPasswordValid(e string) bool {
	number := false
	upper := false
	special := false
	for _, c := range e {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		}
	}

	if number && upper && special && len(e) >= 8 {
		return true
	} else {
		return false
	}
}

func checkUserInput(input string) bool {
	for i := 0; i < len(input); i++ {
		if strings.Contains(disallowed, string(input[i])) {
			return false
		}
	}
	return true
}

func nextBillingCycle(input string) string {
	parts := strings.Split(input, "-")
	day, _ = strconv.Atoi(parts[1])
	month, _ = strconv.Atoi(parts[0])
	year, _ = strconv.Atoi(parts[2])

	if day > 28 {
		day = 28
	}
	if month == 12 {
		month = 1
		year = year + 1
	} else {
		month = month + 1
	}

	return fmt.Sprintf("%d-%d-%d", month, day, year)
}
