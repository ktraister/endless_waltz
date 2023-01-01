package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
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

func checkPrimeNumber(num *big.Int) bool {
	//extend to check above artificial "floor" value
	//perform 20 tests to see if a value is prime or not
	if num.ProbablyPrime(20) {
		fmt.Println("Number is probably prime")
		return true
	} else {
	    fmt.Println("Number is probably not prime...")
	    return false
        }
}

func makeGenerator(prime *big.Int) int {
        //add some logic to this
	return 2
}

func checkGenerator(prime *big.Int, generator int) bool {
	return true
}

func checkPrivKey(key string) bool {
	return true
}

func dh_handshake(conn net.Conn, conn_type string) (string, error) {

	prime := new(big.Int)
	tempkey := new(big.Int)
        tempfloat := new(big.Float)

	var generator int
	var err error
	var ok bool
	buf := make([]byte, 10000)

	if conn_type == "server" {
	        //prime will need to be *big.Int, int cant store the number 
		//possible gen values 2047,3071,4095, 6143, 7679, 8191
		prime, err = rand.Prime(rand.Reader, 2047)
		if err != nil {
			fmt.Println(err)
		}

		//calculate generator
		generator = makeGenerator(prime)

		fmt.Println("Server DH Prime: ", prime)
                fmt.Println("Server DH Generator: ", generator)

		//send the values across the conn
		n, err := conn.Write([]byte(fmt.Sprintf("%d:%d\n", prime, generator)))
		if err != nil {
			log.Println(n, err)
			return "", err
		}
	} else {
		//wait to receive values
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(n, err)
			return "", err
		}
		values := strings.Split(string(buf[:n]), ":")

		prime, ok = prime.SetString(values[0], 0)
		if !ok {
			log.Println("Couldn't convert response prime to int")
			return "", err
		}
		generator, err = strconv.Atoi(strings.Trim(values[1], "\n"))
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		fmt.Println("Client DH Prime: ", prime)
                fmt.Println("Client DH Generator: ", generator)

		//approve the values or bounce the conn
		if checkPrimeNumber(prime) == false || checkGenerator(prime, generator) == false {
		        return "", err
		}
	}

	//myint is private, < p, > 0
	//need to change the method we use here, too
	myint, err := rand.Int(rand.Reader, prime)
	if err != nil {
		log.Println(err)
		return "", err
	}
        tempfloat, ok = tempfloat.SetString(fmt.Sprintf("%s",myint))
		if !ok {
			log.Println("Couldn't convert response tempPubKey to int")
			err = fmt.Errorf("Couldn't convert response tempPubKey to int")
			return "", err
		}
 
        myfloat, accuracy := tempfloat.Float64()
	fmt.Println(accuracy)

	fmt.Println("Private Float: ", myfloat)

	//mod and exchange values
	//compute pubkeys A and B - E.X.) A = g^a mod p : 102 mod 541 = 100
	gofa := math.Pow(float64(generator), myfloat)
	fmt.Println("DEBUG gofa: ", gofa)
	//*** the pubkey we're sending is currently busted***
	pubkey := fmt.Sprintf("%f", math.Mod(gofa, float64(prime.Int64())))

	//clear the buffer
        buf = make([]byte, 10000)

	if conn_type == "server" {
		//send the pubkey across the conn
		fmt.Println("Sending pubkey to client: ", pubkey)
		n, err := conn.Write([]byte(pubkey))
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		n, err = conn.Read(buf)
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		//this needs to be replaced with bigint method setstring
		/*
		tempkey, err = strconv.Atoi(string(buf[:n]))
		if err != nil {
			log.Println(n, err)
			return "", err
		}
		*/

		fmt.Println("Server TempKey: ", string(buf[:n]))
		tempkey, ok = tempkey.SetString(string(buf[:n]), 0)
		if !ok {
			log.Println("Couldn't convert response tempPubKey to int")
                        err = fmt.Errorf("Couldn't convert response tempPubKey to int")
			return "", err
		}

	} else {

		n, err := conn.Read(buf)
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		//send the pubkey across the conn
		n, err = conn.Write([]byte(pubkey))
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		//this needs to be replaced with bigint method setstring
		/*
		tempkey, err = strconv.Atoi(string(buf[:n]))
		if err != nil {
			log.Println(n, err)
			return "", err
		}
		*/
                tempkey, ok = tempkey.SetString(string(buf[:n]), 0)
		if !ok {
			log.Println("Couldn't convert response tempPubKey to int")
			err = fmt.Errorf("Couldn't convert response tempPubKey to int")
			return "", err
		}
	}

	//mod pubkey again E.X.) keya = B^a mod p : 2622 mod 541 = 478
	bofp := math.Pow(float64(tempkey.Int64()), myfloat)
	privkey := fmt.Sprintf("%f", math.Mod(bofp, float64(prime.Int64())))

	if checkPrivKey(privkey) == false {
		// bounce the conn
		return "", err
	}

	//return common secret
	return privkey, nil
}
