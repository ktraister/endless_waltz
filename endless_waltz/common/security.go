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
	var d, m, y int
	parts := strings.Split(input, "-")
	d, _ = strconv.Atoi(parts[1])
	m, _ = strconv.Atoi(parts[0])
	y, _ = strconv.Atoi(parts[2])

	if d > 28 {
		d = 28
	}
	if m == 12 {
		m = 1
		fmt.Println(y)
		y = y + 1
		fmt.Println(y)
	} else {
		m = m + 1
	}

	var month, day string

	//ensure we return date in 01-02-2006 format
	if m < 10 {
		month = fmt.Sprintf("0%d", m)
	} else {
		month = fmt.Sprint(m)
	}

	if d < 10 {
		day = fmt.Sprintf("0%d", d)
	} else {
		day = fmt.Sprint(d)
	}

	return fmt.Sprintf("%s-%s-%d", month, day, y)
}
