package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"strconv"
	"strings"
)

/*
Great RedHat docs on this subject:
https://www.redhat.com/en/blog/understanding-and-verifying-security-diffie-hellman-parameters

And more on the web:
https://crypto.stackexchange.com/questions/820/how-does-one-calculate-a-primitive-root-for-diffie-hellman
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

func findPrimeFactors(i *big.Int) []int {
}

func makeGenerator(prime *big.Int) int {
        //sample code to flesh out this logic
        //https://www.geeksforgeeks.org/primitive-root-of-a-prime-number-n-modulo-n/
	//read the python example to get whats up

	//add this to calculate primitve roots
	one := big.NewInt(1)
	phi := prime.Sub(prime, one)
	flag := false

	//let's figure out our prime factors and store in a map[]
	phiFactors := findPrimeFactors(phi)

	//we'll return i if we get a hit
	for i := big.NewInt(2); i.Cmp(phi) != 0; i.Add(i, one) {
	    //for each i, we need to test 
	    /*
	    for val in phiFactors {
		#python code
		if power(i, phi // val, prime) == 1
		    flag = true
		    break
            */
        }

	return i.Int64()
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

	var generator int
	var err error
	var ok bool
	buf := make([]byte, 10000)

	if conn_type == "server" {
	        //prime will need to be *big.Int, int cant store the number 
		//possible gen values 2047,3071,4095, 6143, 7679, 8191
		prime, err = rand.Prime(rand.Reader, 19)
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

	//mod and exchange values
	//compute pubkeys A and B - E.X.) A = g^a mod p : 102 mod 541 = 100
        tempkey.Exp(big.NewInt(int64(generator)), myint, nil)
	tempkey.Mod(tempkey, prime)

	//clear the buffer
        buf = make([]byte, 10000)

	if conn_type == "server" {
		//send the pubkey across the conn
		fmt.Println("Sending pubkey to client: ", tempkey)
		n, err := conn.Write([]byte(tempkey.String()))
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

		//send the tempkey across the conn
		n, err = conn.Write([]byte(tempkey.String()))
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
	/*
	bofp := math.Pow(exkey, myfloat)
        fmt.Println("DEBUG bofp: ", bofp)
	privkey := fmt.Sprintf("%f", math.Mod(bofp, primefloat))
	*/
	tempkey.Exp(tempkey, myint, nil)
	tempkey.Mod(tempkey, prime)
	privkey := tempkey.String()

	if checkPrivKey(privkey) == false {
		// bounce the conn
		return "", err
	}

	//return common secret
	return privkey, nil
}
