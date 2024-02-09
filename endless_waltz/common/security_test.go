package main

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	if len(generateToken()) != 128 {
		t.Errorf("Expected length is incorret")
	}
}

func TestIsEmailValid(t *testing.T) {
	if !isEmailValid("foo@bar.com") {
		t.Errorf("Correct Email did not pass valid check")
	}
	if isEmailValid("foobar.com") {
		t.Errorf("incorrect Email passed valid check: %s", "foobar.com")
	}
	if isEmailValid("no") {
		t.Errorf("incorrect Email passed valid check: %s", "no")
	}
	/*
		    if isEmailValid("fuck@you") {
			t.Errorf("incorrect Email passed valid check: %s", "fuck@you")
		    }
	*/
}

func TestIsPasswordValid(t *testing.T) {
	if isPasswordValid("foo@bar.com") {
		t.Errorf("Correct Password did not pass valid check")
	}
	if isPasswordValid("foobar.com") {
		t.Errorf("incorrect Password passed valid check: %s", "foobar.com")
	}
	if isPasswordValid("no1") {
		t.Errorf("incorrect Password passed valid check: %s", "no")
	}
	if isPasswordValid("RuckYou") {
		t.Errorf("incorrect Password passed valid check: %s", "fuck@you")
	}
}

func TestCheckUserInput(t *testing.T) {
	if checkUserInput("foo@bar.com") {
		t.Errorf("bad input passed valid check")
	}
	if checkUserInput("use auth; db.keys.find({})") {
		t.Errorf("bad input passed valid check")
	}
	if !checkUserInput("foobar.com") {
		t.Errorf("good input did not pass valid check")
	}
}

func TestNextBillingCycle(t *testing.T) {
	if nextBillingCycle("01-01-2001") != "02-01-2001" {
		t.Errorf("Billing cycle check failed: Expected %s, Got %s", "01-01-2001", nextBillingCycle("01-01-2001"))
	}
	if nextBillingCycle("01-30-2001") != "02-28-2001" {
		t.Errorf("Billing cycle check failed: Expected %s, Got %s", "01-01-2001", nextBillingCycle("01-30-2001"))
	}
	if nextBillingCycle("12-01-2001") != "01-01-2002" {
		t.Errorf("Billing cycle check failed: Expected %s, Got %s", "01-01-2001", nextBillingCycle("12-01-2001"))
	}
	if nextBillingCycle("12-30-2001") != "01-28-2002" {
		t.Errorf("Billing cycle check failed: Expected %s, Got %s", "01-01-2001", nextBillingCycle("12-01-2001"))
	}
}
