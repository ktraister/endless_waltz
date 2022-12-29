package main

import (
    "fmt"
    "crypto/rand"
    "time"
)

func main() {
    for {
	b := make([]byte, 1600)
	n, err := rand.Read(b)
	fmt.Println(n, err, b)
	time.Sleep(8 * time.Second)
    }
}
