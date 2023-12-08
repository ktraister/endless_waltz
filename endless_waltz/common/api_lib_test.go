package main

import (
	"testing"
)

func TestRateLimit(t *testing.T) {
	if !rateLimit("foo", 5) {
		t.Errorf("premature rate limit 1")
	}
	if rateLimit("foo", 1) {
		t.Errorf("uncaught rate limit 1")
	}
}
