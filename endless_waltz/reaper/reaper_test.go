package main

import (
    "testing"
)

func TestPadEncrypt(t *testing.T) {
    output := pad_encrypt("foo", "abcdefg", "12345")
    expectedOutput := "61725, 160485, 148140"
	if output != expectedOutput {
	    t.Errorf("Expected output(%s) is not same as actual output(%s)", expectedOutput, output)
	}
}
