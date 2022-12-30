package main

import (
	"crypto/rand"
	"fmt"
)

/*
Great RedHat docs on this subject:
https://www.redhat.com/en/blog/understanding-and-verifying-security-diffie-hellman-parameters

g = 10
p = 541

a = 2
A = g^a mod p = 102 mod 541 = 100
b = 4
B = g^b mod p = 104 mod 541 = 262
Alice and Bob exchange A and B in view of Carl
keya = B^a mod p = 2622 mod 541 = 478
keyb = A^B mod p = 1004 mod 541 = 478

*/

func dh_handshake(conn *Net.connection, conn_type string) int {

	//calculate my prime and shared prime
	prime, err := rand.Prime(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
	}

	//calculate shared int

	//exchange values

	//mod my prime again

	//return common secret
}
