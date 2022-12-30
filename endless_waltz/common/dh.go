package main

import (
	"fmt"
)

/*
Great RedHat docs on this subject:
https://www.redhat.com/en/blog/understanding-and-verifying-security-diffie-hellman-parameters

//g is a primitive root modulo
g = 10
//p is a shared prime
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
	//myint is private, no more than 1mil, no less than 10thou
	myint := math.rand.Intn(1000000) + 10000
	fmt.Println(myint)

	prime, err := crypto.rand.Prime(crypto.rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
	}

	//calculate shared int
	//sharedint is public, no more than 1mil, no less than 10thou
	sharedint := math.rand.Intn(1000000) + 10000
	fmt.Println(sharedint)

	//exchange values

	//mod my prime again

	//return common secret
}
