package main

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

func dh_handshake(conn *Net.connection) int {
	//calculate my prime

	//calculate shared int

	//exchange values

	//mod my prime again

	//return common secret
}
