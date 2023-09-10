package main

import (
	"bufio"
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

func checkDHPair(num *big.Int, gen int) bool {
	for index, _ := range moduli_pairs {
		values := strings.Split(moduli_pairs[index], ":")
		generator := strconv.Itoa(gen)
		if generator == values[0] && num.String() == values[1] {
			return true
		}
	}
	return false
}

func fetchValues() (*big.Int, int) {
	randomNumber, _ := rand.Int(rand.Reader, big.NewInt(int64(len(moduli_pairs))))
	index := int(randomNumber.Int64())
	values := strings.Split(moduli_pairs[index], ":")
	mod := new(big.Int)
	mod, _ = mod.SetString(values[1], 10)
	gen, _ := strconv.Atoi(values[0])

	return mod, gen
}

func checkPrivKey(key string) bool {
	return true
}

func dh_handshake(conn net.Conn, conn_type string) (string, error) {
	//prime := big.NewInt(1)
	prime := big.NewInt(424889)
	tempkey := big.NewInt(1)

	var generator int
	var err error
	var ok bool

	switch {
	case conn_type == "server":
		//prime will need to be *big.Int, int cant store the number
		//possible gen values 2047,3071,4095, 6143, 7679, 8191

		//replace this with reading from list
		//prime, err = rand.Prime(rand.Reader, 19)
		prime, generator = fetchValues()

		log.Println("Server DH Prime:", prime)
		log.Println("Server DH Generator: ", generator)

		//send the values across the conn
		n, err := conn.Write([]byte(fmt.Sprintf("%d:%d\n", prime, generator)))
		if err != nil {
			log.Println(n, err)
			return "", err
		}
	default:
		//wait to receive values
		var data string

		reader := bufio.NewReader(conn)
		// Read until a newline character is encountered
		data, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		values := strings.Split(data, ":")

		prime, ok = prime.SetString(values[0], 0)
		if !ok {
			log.Println("Couldn't convert response prime to int")
			return "", err
		}
		generator, err = strconv.Atoi(strings.Trim(values[1], "\n"))
		if err != nil {
			log.Println(err)
			return "", err
		}

		log.Println("Client DH Prime: ", prime)
		log.Println("Client DH Generator: ", generator)

		//approve the values or bounce the conn
		if checkDHPair(prime, generator) == false {
			return "", err
		} else {
			log.Println("DH values approved...")
		}
	}

	// SOMETHING ABOUT THIS GENERATION IS BREAKING THE DH SOMETIMES
	// it's something about the generation of the private secrets. I'll come back to it

	//myint is private, < p, > 0
	//need to change the method we use here, too
	//code will now generate an int between 1001 and 2001
	myint, err := rand.Int(rand.Reader, big.NewInt(1001))
	if err != nil {
		log.Println(err)
		return "", err
	}
	myint.Add(myint, big.NewInt(1000))

	//mod and exchange values
	//compute pubkeys A and B - E.X.) A = g^a mod p : 102 mod 541 = 100
	tempkey.Exp(big.NewInt(int64(generator)), myint, nil)
	tempkey.Mod(tempkey, prime)

	switch {
	case conn_type == "server":
		//send the pubkey across the conn
		log.Println("Sending pubkey TO client: ", tempkey)
		n, err := conn.Write([]byte(fmt.Sprintf("%s\n", tempkey.String())))
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		var data string

		reader := bufio.NewReader(conn)
		data, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		data = strings.Replace(data, "\n", "", -1)

		log.Println("Received pubkey FROM client: ", data)
		tempkey, ok = tempkey.SetString(data, 0)
		if !ok {
			log.Println("Couldn't convert response tempPubKey to int")
			err = fmt.Errorf("Couldn't convert response tempPubKey to int")
			return "", err
		}
	default:
		var data string
		reader := bufio.NewReader(conn)
		data, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		data = strings.Replace(data, "\n", "", -1)

		log.Println("Received pubkey FROM server: ", data)

		//send the tempkey across the conn
		log.Println("Sending pubkey TO server: ", tempkey)
		n, err := conn.Write([]byte(fmt.Sprintf("%s\n", tempkey.String())))
		if err != nil {
			log.Println(n, err)
			return "", err
		}

		tempkey, ok = tempkey.SetString(data, 0)
		if !ok {
			log.Println("Couldn't convert response tempPubKey to int: ", data)
			err = fmt.Errorf("Couldn't convert response tempPubKey to int: %s", data)
			return "", err
		}

	}

	tempkey.Exp(tempkey, myint, nil)
	tempkey.Mod(tempkey, prime)
	privkey := tempkey.String()

	if checkPrivKey(privkey) == false {
		// bounce the conn
		return "", err
	}

	//return main secret
	return privkey, nil
}
