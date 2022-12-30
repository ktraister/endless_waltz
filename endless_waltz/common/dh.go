package main

import (
	"fmt"
	"log"
	"math"
	"strings"
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

/*
 * Server should set the Prime and modulo
 * server should also have urandom seeded to garuntee more true randomness :)
 */

func checkPrimeNumber(num int) bool {
	//extend to check above artificial "floor" value
	if num < 2 {
		fmt.Println("Number must be greater than 2.")
		return false
	}
	sq_root := int(math.Sqrt(float64(num)))
	for i := 2; i <= sq_root; i++ {
		if num%i == 0 {
			fmt.Println("Non Prime Number")
			return false
		}
	}
	fmt.Println("Prime Number")
	return true
}

func checkSharedInt(num int) bool {
	return true
}

func dh_handshake(conn *Net.connection, conn_type string) int {

	//calculate my prime and shared prime
	//myint is private, no more than 1mil, no less than 10thou
	myint := math.rand.Intn(1000000) + 10000
	fmt.Println(myint)

	if conn_type == "server" {
		prime, err := crypto.rand.Prime(crypto.rand.Reader, 2048)
		if err != nil {
			fmt.Println(err)
		}

		//calculate shared int
		//sharedint is public, no more than 1mil, no less than 10thou
		sharedint := math.rand.Intn(1000000) + 10000
		fmt.Println(sharedint)

		//send the values across the conn
		n, err := conn.Write([]byte(fmt.Sprintf("%d:%d\n", prime, sharedint)))
		if err != nil {
			log.Println(n, err)
			return
		}
	} else {
		//wait to receive values
		buf := make([]byte, 100)
		conn.Read(buf)
		values := strings.Split(string(buf[:n]), ":")

		//approve the values or bounce the conn
		if checkPrimeNumber(values[0]) == false || checkSharedInt(values[1]) == false {
			//bounce the conn
		}
	}

	//mod and exchange values

	if conn_type == "server" {
	} else {
	}

	//mod my prime again

	//return common secret
}
