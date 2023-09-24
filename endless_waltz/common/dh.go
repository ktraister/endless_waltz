package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
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
	mod, _ = mod.SetString(values[1], 0)
	gen, _ := strconv.Atoi(values[0])

	return mod, gen
}

func checkPrivKey(key string) bool {
	return true
}

func dh_handshake(cm *ConnectionManager, logger *logrus.Logger, configuration Configurations, conn_type string, user string) (string, error) {
	//prime := big.NewInt(1)
	prime := big.NewInt(424889)
	tempkey := big.NewInt(1)

	var generator int
	var err error
	var ok bool

	switch {
	case conn_type == "server":
		//grab a random dh pair from rn.go
		prime, generator = fetchValues()

		logger.Debug("Server DH Prime:", prime)
		logger.Debug("Server DH Generator: ", generator)

		outgoing := &Message{Type: "DH",
			User: configuration.Server.User,
			From: configuration.Server.User,
			To:   user,
			Msg:  fmt.Sprintf("%d:%d", prime, generator),
		}
		b, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		logger.Debug(fmt.Sprintf("Server sending dh pair %s", b))
		err = cm.Send(b)
		if err != nil {
			logger.Fatal("Unable to write message to websocket: ", err)
			return "", err
		}
	default:
		//read in response from server
		_, incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			return "", err
		}

		err = json.Unmarshal([]byte(incoming), &dat)
		if err != nil {
			logger.Error("Error unmarshalling json:", err)
			return "", err
		}

		values := strings.Split(dat["msg"].(string), ":")

		prime, ok = prime.SetString(values[0], 0)
		if !ok {
			logger.Error(fmt.Sprintf("Couldn't convert response prime %s to bigInt", values[0]))
			return "", err
		}
		generator, err = strconv.Atoi(strings.Trim(values[1], "\n"))
		if err != nil {
			logger.Error(err)
			return "", err
		}

		logger.Debug("Client DH Prime: ", prime)
		logger.Debug("Client DH Generator: ", generator)

		//approve the values or bounce the conn
		if checkDHPair(prime, generator) == false {
			logger.Error(err)
			return "", err
		} else {
			logger.Info("DH values approved!")
		}
	}

	/*
		I reaaaallllyyyyyy need to revisit the creation of the private key int
		I understand the THEORY says that 2 <= int < Prime, but huge keys are slow
		And do they even grant extra security? I really don't know.

		Anyway, we'll be revisiting this part of the code many times.
	*/

	//myint is private, int < p, int >= 2
	myint, err := rand.Int(rand.Reader, big.NewInt(9999))
	logger.Debug(fmt.Sprintf("%s chose private int %s", conn_type, myint.String()))
	if err != nil {
		logger.Error(err)
		return "", err
	}
	two := big.NewInt(2)
	if myint.Cmp(two) <= 0 {
		myint.Add(myint, big.NewInt(2))
	}

	//changing base to get some kind of speed boost or something
	prime.Text(2)
	myint.Text(2)
	tempkey.Text(2)

	//mod and exchange values
	//compute pubkeys A and B - E.X.) A = g^a mod p : 102 mod 541 = 100
	tempkey.Exp(big.NewInt(int64(generator)), myint, nil).Mod(tempkey, prime)

	switch {
	case conn_type == "server":
		//send the pubkey across the conn
		logger.Debug("Sending pubkey TO client: ", tempkey)
		outgoing := &Message{Type: "DH",
			User: configuration.Server.User,
			From: configuration.Server.User,
			To:   user,
			Msg:  tempkey.String(),
		}
		b, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		err = cm.Send(b)
		if err != nil {
			logger.Fatal("Unable to write message to websocket: ", err)
			return "", err
		}

		//get client pubkey
		_, incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			return "", err
		}

		err = json.Unmarshal([]byte(incoming), &dat)
		if err != nil {
			logger.Error("Error unmarshalling json:", err)
			return "", err
		}

		logger.Debug("Received pubkey FROM client: ", dat["msg"])
		tempkey, ok = tempkey.SetString(dat["msg"].(string), 0)
		if !ok {
			logger.Error("Couldn't convert response tempPubKey to int")
			err = fmt.Errorf("Couldn't convert response tempPubKey to int")
			return "", err
		}
	default:
		//get client pubkey
		_, incoming, err := cm.Read()
		if err != nil {
			logger.Error("Error reading message:", err)
			return "", err
		}

		err = json.Unmarshal([]byte(incoming), &dat)
		if err != nil {
			logger.Error("Error unmarshalling json:", err)
			return "", err
		}

		logger.Debug("Received pubkey FROM server: ", dat["msg"])

		//send the tempkey across the conn
		logger.Debug("Sending pubkey TO server: ", tempkey.String())
		outgoing := &Message{Type: "DH",
			User: configuration.Server.User,
			From: configuration.Server.User,
			To:   user,
			Msg:  tempkey.String(),
		}
		b, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		err = cm.Send(b)
		if err != nil {
			logger.Fatal("Unable to write message to websocket: ", err)
			return "", err
		}

		tempkey, ok = tempkey.SetString(dat["msg"].(string), 0)
		if !ok {
			logger.Error("Couldn't convert response tempPubKey to int: ", dat["msg"])
			err = fmt.Errorf("Couldn't convert response tempPubKey to int: %s", dat["msg"])
			return "", err
		}

	}

	tempkey.Exp(tempkey, myint, nil).Mod(tempkey, prime)
	//tempkey.Mod(tempkey, prime)
	privkey := tempkey.String()

	if checkPrivKey(privkey) == false {
		// bounce the conn
		return "", err
	}

	//return main secret
	return privkey, nil
}
