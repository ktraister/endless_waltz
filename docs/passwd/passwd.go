package main

import (
    "crypto/sha512"
    "encoding/hex"
    "fmt"
    "os"
)

func main() {
    inputString := os.Args[1]

    // Create a new SHA-512 hash
    hash := sha512.New()

    // Write the string data to the hash
    hash.Write([]byte(inputString))

    // Get the hash sum as a byte slice
    hashSum := hash.Sum(nil)

    // Convert the byte slice to a hexadecimal string
    hashString := hex.EncodeToString(hashSum)

    fmt.Printf("Input String: %s\n", inputString)
    fmt.Printf("SHA-512 Hash: %s\n", hashString)
}
