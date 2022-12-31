package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"crypto/rand"
)

/*
Great RedHat docs on this subject:
https://www.redhat.com/en/blog/understanding-and-verifying-security-diffie-hellman-parameters
/etc/ssh/moduli is helpful too

//g is a primitive root modulo, and generator of p
//when raised to positive whole numbers less than p, never produces the same result
//g is usually a small value
g = 10
//p is a shared prime
p = 541

//both privake keys
//private init keys are both less than p; > 0
a = 2
b = 4

//compute pubkeys A and B
A = g^a mod p : 102 mod 541 = 100
B = g^b mod p : 104 mod 541 = 262

Alice and Bob exchange A and B in view of Carl
keya = B^a mod p : 2622 mod 541 = 478
keyb = A^B mod p : 1004 mod 541 = 478

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

func makeGenerator(num int) int {
	return 0
}

func checkGenerator(num int) bool {
	return true
}

func checkPrivKey(num int) bool {
	return true
}

func dh_handshake(conn net.Conn, conn_type string) int {

        var prime, generator int
	buf := make([]byte, 100)

	if conn_type == "server" {
		//possible gen values 2047,3071,4095, 6143, 7679, 8191
		prime, err := rand.Prime(rand.Reader, 2047)
		if err != nil {
			fmt.Println(err)
		}

		//calculate generator
		generator = makeGenerator(prime)
		fmt.Println(generator)

		//send the values across the conn
		n, err := conn.Write([]byte(fmt.Sprintf("%d:%d\n", prime, generator)))
		if err != nil {
			log.Println(n, err)
			return 0
		}
	} else {
		//wait to receive values
		conn.Read(buf)
		values := strings.Split(string(buf[:n]), ":")

		prime = int(values[0])
		generator = (values[1])

		//approve the values or bounce the conn
		if checkPrimeNumber(prime) == false || checkGenerator(generator) == false {
			//bounce the conn
		}
	}

	//myint is private, < p, > 0
	myint := rand.Int(rand.Reader, prime-1)
	fmt.Println(myint)

	//mod and exchange values
	//compute pubkeys A and B - E.X.) A = g^a mod p : 102 mod 541 = 100
	pubkey := int(math.Pow(generator, myint)) % prime

	if conn_type == "server" {
		//send the pubkey across the conn
		n, err := conn.Write([]byte(pubkey))
		if err != nil {
			log.Println(n, err)
			return 0
		}

		conn.Read(buf)
	} else {

		conn.Read(buf)

		//send the pubkey across the conn
		n, err := conn.Write([]byte(pubkey))
		if err != nil {
			log.Println(n, err)
			return 0
		}
	}

	//mod pubkey again E.X.) keya = B^a mod p : 2622 mod 541 = 478
	privkey := int(math.Pow(string(buf[:n]), myint)) % prime

	if checkPrivKey(privkey) == false {
		// bounce the conn
	}

	//return common secret
	return privkey
}
