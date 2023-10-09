package main

import (
	"fmt"
	"math/rand"
	"time"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWQYZ1234567890"

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
