package main

import (
	"fmt"
	"math/big"
	"strings"
)

/*
         // character to ASCII
         char := 'a' // rune, not string
         ascii := int(char)
         fmt.Println(string(char), " : ", ascii)

         // ASCII to character

         asciiNum := 65  // Uppercase A
         character := string(asciiNum)
         fmt.Println(asciiNum, " : ", character)

	 //outputs
	 a:97
	 65:A
*/

//both these functions need to pass around strings for ease of use

func toString(INPUT []string) string {
	st := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(INPUT)), ", "), "[]")
	return st
}

func fromString(INPUT string) []string {
	trim := strings.ReplaceAll(INPUT, "\n", "")
	s := strings.Split(strings.ReplaceAll(trim, " ", ""), ",")
	mySlice := []string{}
	for _, val := range s {
		mySlice = append(mySlice, val)
	}
	return mySlice
}

func pad_encrypt(MSG string, PAD string, PRIVKEY string) string {
	chars := []rune(MSG)
	pad := []rune(PAD)
	asc_chars := make([]int, 0)
	asc_pad := make([]int, 0)
	enc_msg := make([]string, 0)
	tmpBigInt := big.NewInt(1)
	PrivKeyInt := big.NewInt(1)
	PrivKeyInt.SetString(PRIVKEY, 10)

	//change chars to ascii_chars
	for i := 0; i < len(chars); i++ {
		asc_chars = append(asc_chars, int(chars[i]))
	}

	//change pad to ascii_pad
	for i := 0; i < len(pad); i++ {
		asc_pad = append(asc_pad, int(pad[i]))
	}

	//encode the message
	for i := 0; i < len(asc_chars); i++ {
		//if chars - pad < 255
		val := asc_chars[i] - asc_pad[i]
		if val < 0 {
			val = val + 255
		}
		//operate on the message with PRIVKEY After subtract
		tmpBigInt.SetInt64(int64(val))
		tmpBigInt.Mul(tmpBigInt, PrivKeyInt)
		enc_msg = append(enc_msg, tmpBigInt.String())
	}

	return toString(enc_msg)
}

func pad_decrypt(INPUT_MSG string, PAD string, PRIVKEY string) string {
	pad := []rune(PAD)
	asc_pad := make([]int, 0)
	dec_msg := make([]int, 0)
	tmpBigInt := big.NewInt(1)
	PrivKeyInt := big.NewInt(1)
	PrivKeyInt.SetString(PRIVKEY, 10)

	//convert ENC_MSG string to []int
	ENC_MSG := fromString(INPUT_MSG)

	//change pad to ascii_pad
	for i := 0; i < len(pad); i++ {
		asc_pad = append(asc_pad, int(pad[i]))
	}

	//decrypt message
	for i := 0; i < len(ENC_MSG); i++ {
		//operate on the message with PRIVKEY before add
		tmpBigInt.SetString(ENC_MSG[i], 10)
		tmpBigInt.Div(tmpBigInt, PrivKeyInt)

		//if msg + pad > 255
		val := int(tmpBigInt.Uint64()) + asc_pad[i]
		if val > 255 {
			val = val - 255
		}
		//operate on the message with PRIVKEY
		dec_msg = append(dec_msg, val)
	}

	//change ascii_chars to chars and stringify
	//easy way to do this in go
	//https://stackoverflow.com/questions/40310333/how-to-append-a-character-to-a-string-in-golang
	var sb strings.Builder
	for i := 0; i < len(dec_msg); i++ {
		sb.WriteString(string(rune(dec_msg[i])))
	}
	dec_string := sb.String()

	return dec_string
}
