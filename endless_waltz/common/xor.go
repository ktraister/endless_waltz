package main

import (
	"fmt"
	"strconv"
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

func toString(INPUT []int) string {
	st := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(INPUT)), ", "), "[]")
	return st
}

func fromString(INPUT string) []int {
	trim := strings.ReplaceAll(INPUT, "\n", "")
	s := strings.Split(strings.ReplaceAll(trim, " ", ""), ",")
	mySlice := []int{}
	for _, val := range s {
		myInt, _ := strconv.Atoi(val)
		mySlice = append(mySlice, myInt)
	}
	return mySlice
}

func pad_encrypt(MSG string, PAD string) string {
	chars := []rune(MSG)
	pad := []rune(PAD)
	asc_chars := make([]int, 0)
	asc_pad := make([]int, 0)
	enc_msg := make([]int, 0)

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
		if asc_chars[i] < asc_pad[i] {
			asc_chars[i] = asc_chars[i] + 255
		}
		enc_msg = append(enc_msg, asc_chars[i]-asc_pad[i])
	}

	return toString(enc_msg)
}

func pad_decrypt(INPUT_MSG string, PAD string) string {
	pad := []rune(PAD)
	asc_pad := make([]int, 0)
	dec_msg := make([]int, 0)

	//convert ENC_MSG string to []int
	ENC_MSG := fromString(INPUT_MSG)

	//change pad to ascii_pad
	for i := 0; i < len(pad); i++ {
		asc_pad = append(asc_pad, int(pad[i]))
	}

	//decrypt message
	for i := 0; i < len(ENC_MSG); i++ {
		if ENC_MSG[i] > asc_pad[i] {
			ENC_MSG[i] = ENC_MSG[i] - 255
		}
		dec_msg = append(dec_msg, ENC_MSG[i]+asc_pad[i])
	}

	//change ascii_chars to chars and stringify
	//easy way to do this in go
	//https://stackoverflow.com/questions/40310333/how-to-append-a-character-to-a-string-in-golang
	var sb strings.Builder
	for i := 0; i < len(dec_msg); i++ {
		sb.WriteString(string(dec_msg[i]))
	}
	dec_string := sb.String()

	return dec_string
}

/*
	 main() {
	enc_msg := pad_encrypt("foo", "abcdefg")
	fmt.Println(enc_msg)
	fmt.Println("-------------------------------------")
	dec_msg := pad_decrypt(enc_msg, "abcdefg")
    fmt.Println(dec_msg)
}
*/
